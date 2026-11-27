------------------------- MODULE HandlerRegistration -------------------------
EXTENDS Naturals, FiniteSets, TLC
\* Controller.AddDirective and addHandlerLocked snapshot under c.mtx, check
\* hnd.rel without the lock, and invoke callHandlerUnlocked outside c.mtx.
\* Eligibility and invocation are separate: retirement may occur between them.
\* handler.handleDirective supplies origin context. Error/panic returns skip
\* resolver construction; a successful call attaches only under c.mtx after
\* rechecking registration and directive lifetime. Two calls cover both entry
\* paths and callback reentrancy while retire walks the directive collection.
\* handler.retire publishes rel before cancellation can unlock for callbacks.
\* Each CancelResolver action includes that resolver's removal and cancellation
\* before its callbacks, then allows arbitrary other actions between resolvers.
\* ResolverLifecycle.tla covers value caps, startup and idle notification order.
\* Resolver execution, callback payloads and repeated registration with a new
\* identity are outside this registration's invocation/retirement contract.
Calls == {1, 2}
VARIABLES listed, retired, removal, remaining, call, alive, origin, controllerID,
          attached, detached, canceled, late
vars == <<listed, retired, removal, remaining, call, alive, origin, controllerID,
          attached, detached, canceled, late>>
Init ==
  /\ listed = TRUE /\ retired = FALSE /\ removal = "idle"
  /\ remaining = {}
  /\ call = [i \in Calls |-> "new"]
  /\ alive = Calls
  /\ origin = [i \in Calls |-> "unset"]
  /\ controllerID \in {"", "registration"}
  /\ attached = {} /\ detached = {} /\ canceled = {} /\ late = FALSE
ObserveEligible(i) ==
  /\ call[i] = "new"
  /\ call' = [call EXCEPT ![i] = IF retired THEN "skipped" ELSE "eligible"]
  /\ UNCHANGED <<listed, retired, removal, remaining, alive, origin, controllerID,
                  attached, detached, canceled, late>>
Invoke(i) ==
  /\ call[i] = "eligible"
  /\ call' = [call EXCEPT ![i] = "invoking"]
  /\ origin' = [origin EXCEPT ![i] = IF controllerID = "" THEN "upstream" ELSE controllerID]
  /\ UNCHANGED <<listed, retired, removal, remaining, alive, controllerID,
                  attached, detached, canceled, late>>
Return(i, outcome) ==
  /\ call[i] = "invoking"
  /\ call' = [call EXCEPT ![i] = outcome]
  /\ UNCHANGED <<listed, retired, removal, remaining, alive, origin, controllerID,
                  attached, detached, canceled, late>>
Attach(i) ==
  LET accepted == ~retired /\ i \in alive IN
  /\ call[i] = "returned"
  /\ call' = [call EXCEPT ![i] = "settled"]
  /\ attached' = IF accepted THEN attached \cup {i} ELSE attached
  /\ late' = (late \/ (accepted /\ (retired \/ i \notin alive)))
  /\ UNCHANGED <<listed, retired, removal, remaining, alive, origin, controllerID,
                  detached, canceled>>
RemoveHandler ==
  /\ removal = "idle"
  /\ listed' = FALSE
  /\ retired' = TRUE
  /\ removal' = "canceling"
  /\ remaining' = Calls
  /\ UNCHANGED <<call, alive, origin, controllerID, attached, detached, canceled, late>>
CancelResolver(i) ==
  /\ removal = "canceling" /\ i \in remaining
  /\ remaining' = remaining \ {i}
  /\ detached' = detached \cup (attached \cap {i})
  /\ canceled' = canceled \cup (attached \cap {i})
  /\ attached' = attached \ {i}
  /\ UNCHANGED <<listed, retired, removal, call, alive, origin, controllerID, late>>
FinishRetirement ==
  /\ removal = "canceling" /\ remaining = {}
  /\ removal' = "done"
  /\ UNCHANGED <<listed, retired, remaining, call, alive, origin, controllerID,
                  attached, detached, canceled, late>>
RemoveDirective(i) ==
  /\ i \in alive
  /\ alive' = alive \ {i}
  /\ detached' = detached \cup (attached \cap {i})
  /\ canceled' = canceled \cup (attached \cap {i})
  /\ attached' = attached \ {i}
  /\ UNCHANGED <<listed, retired, removal, remaining, call, origin, controllerID, late>>
Done ==
  /\ removal = "done"
  /\ \A i \in Calls : call[i] \in {"skipped", "error", "settled"}
  /\ UNCHANGED vars
Next == RemoveHandler \/ FinishRetirement \/ Done
        \/ (\E i \in Calls : ObserveEligible(i) \/ Invoke(i) \/ Attach(i)
             \/ Return(i, "returned") \/ Return(i, "error")
             \/ CancelResolver(i) \/ RemoveDirective(i))
Spec == Init /\ [][Next]_vars
RetirementBeforeCallbacks == removal # "idle" => (~listed /\ retired)
InvocationRetirementOrder == ~late
RetirementComplete == removal = "done" => attached = {}
CancellationBeforeRelease == detached \subseteq canceled
OriginContext == \A i \in Calls : origin[i] # "unset" =>
  origin[i] = IF controllerID = "" THEN "upstream" ELSE controllerID
=============================================================================
