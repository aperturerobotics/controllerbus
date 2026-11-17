---- MODULE ResolverLifecycle ----
EXTENDS TLC
\* One HandleDirective call, from AddDirective or AddHandler, returns a resolver
\* outside c.mtx. Attachment and handler/directive removal hold c.mtx. A full
\* directive keeps an attached resolver pending until handleNotFullLocked.
\* updateContextLocked publishes the context before notifying idle callbacks;
\* a callback can remove the handler or directive while c.mtx is unlocked.
\* Other controllers, values, and callback payloads do not change these guards.
VARIABLES handler, directive, call, resolver, context, notify, full
vars == <<handler, directive, call, resolver, context, notify, full>>
Init ==
  /\ handler = "attached"
  /\ directive = "attached"
  /\ call = "pending"
  /\ resolver = "absent"
  /\ context = "absent"
  /\ notify = FALSE
  /\ full \in BOOLEAN
ReturnHandler ==
  /\ call = "pending"
  /\ call' = "returned"
  /\ UNCHANGED <<handler, directive, resolver, context, notify, full>>
AttachResolver ==
  /\ call = "returned"
  /\ resolver = "absent"
  /\ handler = "attached" /\ directive = "attached"
  /\ resolver' = "pending"
  /\ UNCHANGED <<handler, directive, call, context, notify, full>>
StartResolver ==
  /\ resolver = "pending" /\ ~full
  /\ handler = "attached" /\ directive = "attached"
  /\ resolver' = "started"
  /\ context' = "running"
  /\ notify' = TRUE
  /\ UNCHANGED <<handler, directive, call, full>>
NotifyIdle ==
  /\ notify
  /\ notify' = FALSE
  /\ UNCHANGED <<handler, directive, call, resolver, context, full>>
RemoveHandler ==
  /\ handler = "attached"
  /\ handler' = "removed"
  /\ resolver' = IF resolver = "absent" THEN "absent" ELSE "removed"
  /\ context' = IF context = "running" THEN "cancelled" ELSE context
  /\ UNCHANGED <<directive, call, notify, full>>
RemoveDirective ==
  /\ directive = "attached"
  /\ directive' = "removed"
  /\ resolver' = IF resolver = "absent" THEN "absent" ELSE "removed"
  /\ context' = IF context = "running" THEN "cancelled" ELSE context
  /\ UNCHANGED <<handler, call, notify, full>>
DropBelowCap ==
  /\ full
  /\ full' = FALSE
  /\ UNCHANGED <<handler, directive, call, resolver, context, notify>>
Done ==
  /\ call = "returned"
  /\ ~notify /\ ~full
  /\ handler = "removed" /\ directive = "removed"
  /\ UNCHANGED vars
Next == ReturnHandler \/ AttachResolver \/ StartResolver \/ NotifyIdle
        \/ RemoveHandler \/ RemoveDirective \/ DropBelowCap \/ Done
Spec == Init /\ [][Next]_vars
TypeOK ==
  /\ handler \in {"attached", "removed"}
  /\ directive \in {"attached", "removed"}
  /\ call \in {"pending", "returned"}
  /\ resolver \in {"absent", "pending", "started", "removed"}
  /\ context \in {"absent", "running", "cancelled"}
  /\ notify \in BOOLEAN /\ full \in BOOLEAN
OnlyLiveRegistrations ==
  resolver \in {"pending", "started"} =>
    handler = "attached" /\ directive = "attached"
RemovedCannotRun ==
  handler = "removed" \/ directive = "removed" => context # "running"
====
