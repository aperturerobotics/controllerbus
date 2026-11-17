---- MODULE StateNotifications ----
EXTENDS Naturals, Sequences
\* Each Mutate is an outer deferred mutation boundary under c.mtx. Nested
\* mutations are coalesced before that boundary. State callback delivery uses
\* the existing serialized callback queue, with c.mtx unlocked during a call.
\* The two mutations may represent values, resolver errors, idle, or removal;
\* all invalidate the same cached state snapshot. Callback payloads and other
\* directives do not affect this per-instance notification protocol.
CONSTANT HoldSuppressionDuringCallback
VARIABLES revision, delivered, queue, active, suppressed
vars == <<revision, delivered, queue, active, suppressed>>
Init ==
  /\ revision = 0 /\ delivered = 0 /\ queue = <<>>
  /\ active = FALSE /\ suppressed = FALSE
Mutate ==
  /\ revision < 2
  /\ revision' = revision + 1
  /\ queue' = IF suppressed THEN queue ELSE Append(queue, revision + 1)
  /\ suppressed' = HoldSuppressionDuringCallback
  /\ UNCHANGED <<delivered, active>>
Dispatch ==
  /\ ~active /\ Len(queue) > 0
  /\ delivered' = Head(queue)
  /\ queue' = Tail(queue)
  /\ active' = TRUE
  /\ UNCHANGED <<revision, suppressed>>
FinishCallback ==
  /\ active
  /\ active' = FALSE /\ suppressed' = FALSE
  /\ UNCHANGED <<revision, delivered, queue>>
Done ==
  /\ revision = 2 /\ queue = <<>> /\ ~active
  /\ UNCHANGED vars
Next == Mutate \/ Dispatch \/ FinishCallback \/ Done
Spec == Init /\ [][Next]_vars
TypeOK ==
  /\ revision \in 0..2 /\ delivered \in 0..2
  /\ queue \in Seq(0..2)
  /\ active \in BOOLEAN /\ suppressed \in BOOLEAN
QuiescentSeesLatest ==
  revision = 2 /\ queue = <<>> /\ ~active => delivered = revision
====
