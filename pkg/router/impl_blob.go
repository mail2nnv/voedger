/*
 * Copyright (c) 2022-present unTill Pro, Ltd.
 */

package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	ibus "github.com/voedger/voedger/staging/src/github.com/untillpro/airs-ibus"

	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/coreutils"
	"github.com/voedger/voedger/pkg/coreutils/federation"
	"github.com/voedger/voedger/pkg/coreutils/utils"
	"github.com/voedger/voedger/pkg/goutils/logger"
	"github.com/voedger/voedger/pkg/iblobstorage"
	"github.com/voedger/voedger/pkg/iblobstoragestg"
	"github.com/voedger/voedger/pkg/iprocbus"
	"github.com/voedger/voedger/pkg/istructs"
)

type blobWriteDetailsSingle struct {
	name     string
	mimeType string
	duration iblobstorage.DurationType
}

type blobWriteDetailsMultipart struct {
	boundary string
	duration iblobstorage.DurationType
}

type blobReadDetails_Persistent struct {
	blobID istructs.RecordID
}

type blobReadDetails_Temporary struct {
	suuid iblobstorage.SUUID
}

type blobBaseMessage struct {
	req             *http.Request
	resp            http.ResponseWriter
	doneChan        chan struct{}
	wsid            istructs.WSID
	appQName        appdef.AppQName
	header          map[string][]string
	wLimiterFactory func() iblobstorage.WLimiterType
}

type blobMessage struct {
	blobBaseMessage
	// could be blobReadDetails_Temporary or blobReadDetails_Persistent
	blobDetails interface{}
}

func (bm *blobBaseMessage) Release() {
	bm.req.Body.Close()
}

func blobReadMessageHandler(bbm blobBaseMessage, blobReadDetails interface{}, blobStorage iblobstorage.IBLOBStorage, bus ibus.IBus, busTimeout time.Duration) {
	defer close(bbm.doneChan)

	// request to VVM to check the principalToken
	req := ibus.Request{
		Method:   ibus.HTTPMethodPOST,
		WSID:     bbm.wsid,
		AppQName: bbm.appQName.String(),
		Resource: "q.sys.DownloadBLOBAuthnz",
		Header:   bbm.header,
		Body:     []byte(`{}`),
		Host:     localhost,
	}
	blobHelperResp, _, _, err := bus.SendRequest2(bbm.req.Context(), req, busTimeout)
	if err != nil {
		WriteTextResponse(bbm.resp, "failed to exec q.sys.DownloadBLOBAuthnz: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if blobHelperResp.StatusCode != http.StatusOK {
		WriteTextResponse(bbm.resp, "q.sys.DownloadBLOBAuthnz returned error: "+string(blobHelperResp.Data), blobHelperResp.StatusCode)
		return
	}

	// read the BLOB
	var blobKey iblobstorage.IBLOBKey
	switch typedKey := blobReadDetails.(type) {
	case blobReadDetails_Persistent:
		blobKey = &iblobstorage.PersistentBLOBKeyType{
			ClusterAppID: istructs.ClusterAppID_sys_blobber,
			WSID:         bbm.wsid,
			BlobID:       typedKey.blobID,
		}
	case blobReadDetails_Temporary:
		blobKey = &iblobstorage.TempBLOBKeyType{
			ClusterAppID: istructs.ClusterAppID_sys_blobber,
			WSID:         bbm.wsid,
			SUUID:        typedKey.suuid,
		}
	default:
		// notest
		panic(fmt.Sprintf("unexpected blobReadDetails: %T", blobReadDetails))
	}
	stateWriterDiscard := func(state iblobstorage.BLOBState) error {
		if state.Status != iblobstorage.BLOBStatus_Completed {
			return errors.New("blob is not completed")
		}
		if len(state.Error) > 0 {
			return errors.New(state.Error)
		}
		bbm.resp.Header().Set(coreutils.ContentType, state.Descr.MimeType)
		bbm.resp.Header().Add("Content-Disposition", fmt.Sprintf(`attachment;filename="%s"`, state.Descr.Name))
		bbm.resp.WriteHeader(http.StatusOK)
		return nil
	}
	if err := blobStorage.ReadBLOB(bbm.req.Context(), blobKey, stateWriterDiscard, bbm.resp, iblobstoragestg.RLimiter_Null); err != nil {
		logger.Error(fmt.Sprintf("failed to read or send BLOB: id %s, appQName %s, wsid %d: %s", blobKey.ID(), bbm.appQName, bbm.wsid, err.Error()))
		if errors.Is(err, iblobstorage.ErrBLOBNotFound) {
			WriteTextResponse(bbm.resp, err.Error(), http.StatusNotFound)
			return
		}
		WriteTextResponse(bbm.resp, err.Error(), http.StatusInternalServerError)
	}
}

// returns NullRecordID for temporary BLOB
func registerBLOB(ctx context.Context, wsid istructs.WSID, appQName string, registerFuncName string, header map[string][]string, busTimeout time.Duration,
	bus ibus.IBus, resp http.ResponseWriter) (ok bool, blobID istructs.RecordID) {
	req := ibus.Request{
		Method:   ibus.HTTPMethodPOST,
		WSID:     wsid,
		AppQName: appQName,
		Resource: registerFuncName,
		Body:     []byte(`{}`),
		Header:   header,
		Host:     localhost,
	}
	blobHelperResp, _, _, err := bus.SendRequest2(ctx, req, busTimeout)
	if err != nil {
		WriteTextResponse(resp, fmt.Sprintf("failed to exec %s: %s", registerFuncName, err.Error()), http.StatusInternalServerError)
		return false, istructs.NullRecordID
	}
	if blobHelperResp.StatusCode != http.StatusOK {
		WriteTextResponse(resp, fmt.Sprintf("%s returned error: %s", registerFuncName, string(blobHelperResp.Data)), blobHelperResp.StatusCode)
		return false, istructs.NullRecordID
	}

	cmdResp := map[string]interface{}{}
	if err := json.Unmarshal(blobHelperResp.Data, &cmdResp); err != nil {
		WriteTextResponse(resp, fmt.Sprintf("failed to json-unmarshal %s result: %s", registerFuncName, err), http.StatusInternalServerError)
		return false, istructs.NullRecordID
	}
	newIDsIntf, ok := cmdResp["NewIDs"]
	if ok {
		newIDs := newIDsIntf.(map[string]interface{})
		return true, istructs.RecordID(newIDs["1"].(float64))
	}
	return true, istructs.NullRecordID
}

func writeBLOB_temporary(ctx context.Context, wsid istructs.WSID, appQName string, header map[string][]string, resp http.ResponseWriter,
	blobName, blobMimeType string, blobDuration iblobstorage.DurationType, blobStorage iblobstorage.IBLOBStorage, body io.ReadCloser,
	bus ibus.IBus, busTimeout time.Duration, wLimiterFactory func() iblobstorage.WLimiterType) (blobSUUID iblobstorage.SUUID) {
	registerFuncName, ok := durationToRegisterFuncs[blobDuration]
	if !ok {
		// notest
		WriteTextResponse(resp, "unsupported blob duration value: "+strconv.Itoa(int(blobDuration)), http.StatusBadRequest)
		return ""
	}

	if ok, _ = registerBLOB(ctx, wsid, appQName, registerFuncName, header, busTimeout, bus, resp); !ok {
		return
	}

	blobSUUID = iblobstorage.NewSUUID()
	key := iblobstorage.TempBLOBKeyType{
		ClusterAppID: istructs.ClusterAppID_sys_blobber,
		WSID:         wsid,
		SUUID:        blobSUUID,
	}
	descr := iblobstorage.DescrType{
		Name:     blobName,
		MimeType: blobMimeType,
	}

	wLimiter := wLimiterFactory()
	if err := blobStorage.WriteTempBLOB(ctx, key, descr, body, wLimiter, blobDuration); err != nil {
		if errors.Is(err, iblobstorage.ErrBLOBSizeQuotaExceeded) {
			WriteTextResponse(resp, err.Error(), http.StatusForbidden)
			return ""
		}
		WriteTextResponse(resp, err.Error(), http.StatusInternalServerError)
		return ""
	}
	return blobSUUID
}

func writeBLOB_persistent(ctx context.Context, wsid istructs.WSID, appQName string, header map[string][]string, resp http.ResponseWriter,
	blobName, blobMimeType string, blobStorage iblobstorage.IBLOBStorage, body io.ReadCloser,
	bus ibus.IBus, busTimeout time.Duration, wLimiterFactory func() iblobstorage.WLimiterType) (blobID istructs.RecordID) {

	// request VVM for check the principalToken and get a blobID
	ok := false
	if ok, blobID = registerBLOB(ctx, wsid, appQName, "c.sys.UploadBLOBHelper", header, busTimeout, bus, resp); !ok {
		return
	}

	// write the BLOB
	key := iblobstorage.PersistentBLOBKeyType{
		ClusterAppID: istructs.ClusterAppID_sys_blobber,
		WSID:         wsid,
		BlobID:       blobID,
	}
	descr := iblobstorage.DescrType{
		Name:     blobName,
		MimeType: blobMimeType,
	}

	wLimiter := wLimiterFactory()
	if err := blobStorage.WriteBLOB(ctx, key, descr, body, wLimiter); err != nil {
		if errors.Is(err, iblobstorage.ErrBLOBSizeQuotaExceeded) {
			WriteTextResponse(resp, err.Error(), http.StatusForbidden)
			return 0
		}
		WriteTextResponse(resp, err.Error(), http.StatusInternalServerError)
		return 0
	}

	// set WDoc<sys.BLOB>.status = BLOBStatus_Completed
	req := ibus.Request{
		Method:   ibus.HTTPMethodPOST,
		WSID:     wsid,
		AppQName: appQName,
		Resource: "c.sys.CUD",
		Body:     []byte(fmt.Sprintf(`{"cuds":[{"sys.ID": %d,"fields":{"status":%d}}]}`, blobID, iblobstorage.BLOBStatus_Completed)),
		Header:   header,
		Host:     localhost,
	}
	cudWDocBLOBUpdateResp, _, _, err := bus.SendRequest2(ctx, req, busTimeout)
	if err != nil {
		WriteTextResponse(resp, "failed to exec c.sys.CUD: "+err.Error(), http.StatusInternalServerError)
		return 0
	}
	if cudWDocBLOBUpdateResp.StatusCode != http.StatusOK {
		WriteTextResponse(resp, "c.sys.CUD returned error: "+string(cudWDocBLOBUpdateResp.Data), cudWDocBLOBUpdateResp.StatusCode)
		return 0
	}

	return blobID
}

func blobWriteMessageHandlerMultipart(bbm blobBaseMessage, blobStorage iblobstorage.IBLOBStorage, blobDetails blobWriteDetailsMultipart,
	bus ibus.IBus, busTimeout time.Duration) {
	defer close(bbm.doneChan)

	r := multipart.NewReader(bbm.req.Body, blobDetails.boundary)
	var part *multipart.Part
	var err error
	blobIDsOrSUUIDs := []string{}
	partNum := 0
	for err == nil {
		part, err = r.NextPart()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				WriteTextResponse(bbm.resp, "failed to parse multipart: "+err.Error(), http.StatusBadRequest)
				return
			} else if partNum == 0 {
				WriteTextResponse(bbm.resp, "empty multipart request", http.StatusBadRequest)
				return
			}
			break
		}
		contentDisposition := part.Header.Get(coreutils.ContentDisposition)
		mediaType, params, err := mime.ParseMediaType(contentDisposition)
		if err != nil {
			WriteTextResponse(bbm.resp, fmt.Sprintf("failed to parse Content-Disposition of part number %d: %s", partNum, contentDisposition), http.StatusBadRequest)
		}
		if mediaType != "form-data" {
			WriteTextResponse(bbm.resp, fmt.Sprintf("unsupported ContentDisposition mediaType of part number %d: %s", partNum, mediaType), http.StatusBadRequest)
		}
		contentType := part.Header.Get(coreutils.ContentType)
		if len(contentType) == 0 {
			contentType = coreutils.ApplicationXBinary
		}
		part.Header[coreutils.Authorization] = bbm.header[coreutils.Authorization] // add auth header for c.sys.*BLOBHelper
		blobIDOrSUUID := ""
		if blobDetails.duration > 0 {
			// temporary BLOB
			blobIDOrSUUID = string(writeBLOB_temporary(bbm.req.Context(), bbm.wsid, bbm.appQName.String(), part.Header, bbm.resp,
				params["name"], contentType, blobDetails.duration, blobStorage, part, bus, busTimeout, bbm.wLimiterFactory))
		} else {
			// persistent BLOB
			blobID := writeBLOB_persistent(bbm.req.Context(), bbm.wsid, bbm.appQName.String(), part.Header, bbm.resp,
				params["name"], contentType, blobStorage, part, bus, busTimeout, bbm.wLimiterFactory)
			blobIDOrSUUID = utils.UintToString(blobID)
		}
		if len(blobIDOrSUUID) == 0 {
			return // request handled
		}
		blobIDsOrSUUIDs = append(blobIDsOrSUUIDs, blobIDOrSUUID)
		partNum++
	}
	WriteTextResponse(bbm.resp, strings.Join(blobIDsOrSUUIDs, ","), http.StatusOK)
}

func blobWriteMessageHandlerSingle(bbm blobBaseMessage, blobWriteDetails blobWriteDetailsSingle, blobStorage iblobstorage.IBLOBStorage, header map[string][]string,
	bus ibus.IBus, busTimeout time.Duration) {
	defer close(bbm.doneChan)

	if blobWriteDetails.duration > 0 {
		// remporary BLOB
		blobSUUID := writeBLOB_temporary(bbm.req.Context(), bbm.wsid, bbm.appQName.String(), header, bbm.resp, blobWriteDetails.name,
			blobWriteDetails.mimeType, blobWriteDetails.duration, blobStorage, bbm.req.Body, bus, busTimeout, bbm.wLimiterFactory)
		if len(blobSUUID) > 0 {
			WriteTextResponse(bbm.resp, string(blobSUUID), http.StatusOK)
		}
	} else {
		// persistent BLOB
		blobID := writeBLOB_persistent(bbm.req.Context(), bbm.wsid, bbm.appQName.String(), header, bbm.resp, blobWriteDetails.name,
			blobWriteDetails.mimeType, blobStorage, bbm.req.Body, bus, busTimeout, bbm.wLimiterFactory)
		if blobID > 0 {
			WriteTextResponse(bbm.resp, utils.UintToString(blobID), http.StatusOK)
		}
	}
}

// ctx here is VVM context. It used to track VVM shutdown. Blobber will use the request's context
func blobMessageHandler(vvmCtx context.Context, sc iprocbus.ServiceChannel, blobStorage iblobstorage.IBLOBStorage, bus ibus.IBus, busTimeout time.Duration) {
	for vvmCtx.Err() == nil {
		select {
		case mesIntf := <-sc:
			blobMessage := mesIntf.(blobMessage)
			switch blobDetails := blobMessage.blobDetails.(type) {
			case blobReadDetails_Persistent, blobReadDetails_Temporary:
				blobReadMessageHandler(blobMessage.blobBaseMessage, blobDetails, blobStorage, bus, busTimeout)
			case blobWriteDetailsSingle:
				blobWriteMessageHandlerSingle(blobMessage.blobBaseMessage, blobDetails, blobStorage, blobMessage.header, bus, busTimeout)
			case blobWriteDetailsMultipart:
				blobWriteMessageHandlerMultipart(blobMessage.blobBaseMessage, blobStorage, blobDetails, bus, busTimeout)
			}
		case <-vvmCtx.Done():
			return
		}
	}
}

func (s *httpService) blobRequestHandler(resp http.ResponseWriter, req *http.Request, details interface{}) {
	vars := mux.Vars(req)
	wsid, err := strconv.ParseUint(vars[URLPlaceholder_wsid], utils.DecimalBase, utils.BitSize64)
	if err != nil {
		// notest: checked by router url rule
		panic(err)
	}
	headers := maps.Clone(req.Header)
	if _, ok := headers[coreutils.Authorization]; !ok {
		// no token among headers -> look among cookies
		// no token among cookies as well -> just do nothing, 403 will happen on call helper commands further in BLOBs processor
		cookie, err := req.Cookie(coreutils.Authorization)
		if !errors.Is(err, http.ErrNoCookie) {
			if err != nil {
				// notest
				panic(err)
			}
			val, err := url.QueryUnescape(cookie.Value)
			if err != nil {
				WriteTextResponse(resp, "failed to unescape cookie '"+cookie.Value+"'", http.StatusBadRequest)
				return
			}
			// authorization token in cookies -> q.sys.DownloadBLOBAuthnz requires it in headers
			headers[coreutils.Authorization] = []string{val}
		}
	}
	mes := blobMessage{
		blobBaseMessage: blobBaseMessage{
			req:             req,
			resp:            resp,
			wsid:            istructs.WSID(wsid),
			doneChan:        make(chan struct{}),
			appQName:        appdef.NewAppQName(vars[URLPlaceholder_appOwner], vars[URLPlaceholder_appName]),
			header:          headers,
			wLimiterFactory: s.WLimiterFactory,
		},
		blobDetails: details,
	}
	if !s.BlobberParams.procBus.Submit(0, 0, mes) {
		resp.WriteHeader(http.StatusServiceUnavailable)
		resp.Header().Add("Retry-After", strconv.Itoa(s.BlobberParams.RetryAfterSecondsOn503))
		return
	}
	<-mes.doneChan
}

func (s *httpService) blobReadRequestHandler() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		blobIDStr := vars[URLPlaceholder_blobID]
		var blobReadDetails interface{}
		if len(blobIDStr) > temporaryBLOBIDLenTreshold {
			// consider the blobID contains SUUID of a temporary BLOB
			blobReadDetails = blobReadDetails_Temporary{
				suuid: iblobstorage.SUUID(blobIDStr),
			}
		} else {
			// conider the BLOB is persistent
			blobID, err := strconv.ParseUint(blobIDStr, utils.DecimalBase, utils.BitSize64)
			if err != nil {
				// notest: checked by router url rule
				panic(err)
			}
			blobReadDetails = blobReadDetails_Persistent{
				blobID: istructs.RecordID(blobID),
			}
		}
		s.blobRequestHandler(resp, req, blobReadDetails)
	}
}

func (s *httpService) blobWriteRequestHandler() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		queryParamName, queryParamMimeType, boundary, duration, ok := getBlobParams(resp, req)
		if !ok {
			return
		}

		if len(queryParamName) > 0 {
			s.blobRequestHandler(resp, req, blobWriteDetailsSingle{
				name:     queryParamName,
				mimeType: queryParamMimeType,
				duration: duration,
			})
		} else {
			s.blobRequestHandler(resp, req, blobWriteDetailsMultipart{
				boundary: boundary,
				duration: duration,
			})
		}
	}
}

// determines BLOBs write kind: name+mimeType in query params -> single BLOB, body is BLOB content, otherwise -> body is multipart/form-data
// (is multipart/form-data) == len(boundary) > 0
func getBlobParams(rw http.ResponseWriter, req *http.Request) (name, mimeType, boundary string, duration iblobstorage.DurationType, ok bool) {
	badRequest := func(msg string) {
		WriteTextResponse(rw, msg, http.StatusBadRequest)
	}
	values := req.URL.Query()
	nameQuery, isSingleBLOB := values["name"]
	mimeTypeQuery, hasMimeType := values["mimeType"]
	ttlQuery := values["ttl"]
	if (isSingleBLOB && !hasMimeType) || (!isSingleBLOB && hasMimeType) {
		badRequest("both name and mimeType query params must be specified")
		return
	}

	if len(ttlQuery) > 0 {
		// temporary BLOB
		ttl := ttlQuery[0]
		ttlSupported := false
		if duration, ttlSupported = federation.TemporaryBLOB_URLTTLToDurationLs[ttl]; !ttlSupported {
			badRequest(`"1d" is only supported for now for temporary blob ttl`)
			return
		}
	}

	contentType := req.Header.Get(coreutils.ContentType)
	if isSingleBLOB {
		if contentType == "multipart/form-data" {
			badRequest(`name+mimeType query params and "multipart/form-data" Content-Type header are mutual exclusive`)
			return
		}
		name = nameQuery[0]
		mimeType = mimeTypeQuery[0]
		ok = true
		return
	}
	if len(contentType) == 0 {
		badRequest(`neither "name"+"mimeType" query params nor Content-Type header is not provided`)
		return
	}
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		badRequest("failed ot parse Content-Type header: " + contentType)
		return
	}
	if mediaType != "multipart/form-data" {
		badRequest("name+mimeType query params are not provided -> Content-Type must be mutipart/form-data but actual is " + contentType)
		return
	}
	boundary = params["boundary"]
	if len(boundary) == 0 {
		badRequest("boundary of multipart/form-data is not specified")
		return
	}
	return name, mimeType, boundary, duration, true
}
