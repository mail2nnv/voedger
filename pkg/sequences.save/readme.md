
# Sequences

## Introduction

```mermaid
flowchart

subgraph AppParts [App partition]
    Deploy@{ shape: fr-rect, label: "Deploy" }
end

Deploy --> Recovery

subgraph processor [CP event loop]
    InputEvent@{ shape: lean-r, label: "Get raw event" }
    WaitSequences@{ shape: hex, label: "Wait for Sequences recovered" }
    subgraph BuildPLogEvent
      Offsets@{ shape: fr-rect, label: "Calc offsets" }
      CUDs@{ shape: fr-rect, label: "Build CUDs" }
      Args@{ shape: fr-rect, label: "Build Args" }
      Offsets --> CUDs
      CUDs --> Args
    end
    Args --> SavePLogEvent

    fail@{ shape: tri, label: "error" }

    SavePLogEvent@{ shape: hex, label: "Save PLog event" }
    ApplyCUDs@{ shape: fr-rect, label: "Apply event records" }
    SyncProj@{ shape: fr-rect, label: "Sync projectors" }
    SaveWLog@{ shape: fr-rect, label: "Save WLog event" }

        
    InputEvent --> WaitSequences 
    WaitSequences --> Offsets
    WaitSequences -..->|recovery timeout| fail
    SavePLogEvent --> ApplyCUDs
    ApplyCUDs --> SyncProj
    SyncProj --> SaveWLog
end

WaitSequences <-..-> WaitForRecovery

Offsets <-..->|plog, wlog offsets| Next
CUDs <-..->|crec,wrec ids| Next
Args <-..->|orec ids| Next

SavePLogEvent -..->|Success| Apply
SavePLogEvent -..->|Fail| Discard

subgraph Sequencer ["Sequences [PartitionID]"]
  subgraph API
    Recovery@{ shape: fr-rect, label: "Recovery(ctx)" }
    WaitForRecovery@{ shape: fr-rect, label: "WaitForRecovery(Duration) bool" }
    Next@{ shape: fr-rect, label: "Next(WSID, QName)" }
    Apply@{ shape: fr-rect, label: "Apply()" }
    Discard@{ shape: fr-rect, label: "Discard()" }
    Recovery~~~WaitForRecovery~~~Next~~~Discard~~~Apply
  end

  Apply -..->|calls| apply
  Discard -..->|clear| changes
  
  subgraph impl
    changes@{ shape: doc, label: "changes map[wsid,QName]uint64" }
    apply@{ shape: rounded, label: "apply changes" }
    chan@{ shape: h-cyl, label: "chanel[changes]" }
    updateView@{ shape: rounded, label: "⚡ update sequence view" }
    recoveryRoutine@{ shape: rounded, label: "⚡ recovery routine" }
    changes~~~apply~~~chan~~~updateView~~~recoveryRoutine
  end

  Next -..->|collect|changes
  changes -..->|read| apply
  apply -..->|pass changes| chan
  chan -..-> updateView
  Recovery -..->|starts| recoveryRoutine
  recoveryRoutine -..->|waits|WaitForRecovery
end


subgraph Cassandra [Cassandra DB]
  PLog@{ shape: cyl, label: "PLog" }
  SeqView@{ shape: cyl, label: "Sequence view" }
end

recoveryRoutine <-..->|reads & recovers|SeqView
updateView -..->|updates|SeqView

PLog -..->|reads| recoveryRoutine
SavePLogEvent -..->|writes| PLog
```

## Sequence view

### Structure

```mermaid
erDiagram
    SequencesView {
    }
    SequencesView ||--|| Key : key
    Key {
    }
    Key ||--|| PK : PK
    Key ||--|| CC : CC
    PK {
        uint64 pid
    }
    CC {
        uint64 wsid
        QName  name
    }
    SequencesView ||--|| Value : Value
    Value {
        uint64 last
    }
```

### Contens

```mermaid
erDiagram
    SequencesView {
    }
    SequencesView ||--|| PLogOffset : has
    PLogOffset["PLog Offset sequence"] {
      uint64 pid PK
      uint64 wsid PK "0 (PLog)"
      QName  name PK "sys.plogOffsetSeq"
      uint64 last
    }
    SequencesView ||--o{ WorkspaceSequences : has
    WorkspaceSequences["Workspace sequences"] {
    }
    WorkspaceSequences ||--|| WLogOffset : includes
    WLogOffset["WLog Offset sequence"] {
      uint64 pid PK
      uint64 wsid PK
      QName  name PK "sys.wlogOffsetSeq"
      uint64 last
    }
    WorkspaceSequences ||--|| RecordID : includes
    RecordID["WRecords and ORecords ID sequence"] {
      uint64 pid PK
      uint64 wsid PK
      QName  name PK "sys.recIDSeq"
      uint64 last
    }
    WorkspaceSequences ||--|| CRecordID : includes
    CRecordID["CRecords ID sequence"] {
      uint64 pid PK
      uint64 wsid PK
      QName  name PK "sys.cRecIDSeq"
      uint64 last
    }
```

### Sizes

**Overall record size:**

8 + 8 + 2 + 8 = **26** bytes.

**Max records per partition:**

Max Cassandra partition size is 20Mb.

Maximum 20 × 1024 × 1024 / 26 = **806’596** records per partition.

**Max sequences per workspace:**

Let we have 1’000 partitions.
Maximum 806’596 / 1’000= **806** sequences per workspace.

**Max workspaces per partition:**

Let we have only system sequences:

sys.plogOffsetSeq: 1
sys.wlogOffsetSeq, sys.recIDSeq, sys.cRecIDSeq: 3 per ws
Maximum (806’596 - 1) / 3 = **268’865** workspaces per partition.
