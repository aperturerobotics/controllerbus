-------------------- MODULE ReferenceRetirement --------------------
EXTENDS Naturals
CONSTANT DetachBeforeRelease
VARIABLES referenceLock, mapLock, callback, cleanup
vars == <<referenceLock, mapLock, callback, cleanup>>

\* ApplyConfigSet and LookupConfigSet callbacks run under runningControllerRef's
\* mutex. Both resolvers retire references when their directive context ends.
\* Appliers and config lookups are callers of this path.
\* The old cleanup took refsMtx and then Reference.Release's mutex, reversing
\* the callback's order. Detaching the map lets Release wait without refsMtx.
Init == /\ referenceLock = "free" /\ mapLock = "free"
        /\ callback = 0 /\ cleanup = 0

CallbackAcquireReference == /\ callback = 0 /\ referenceLock = "free"
                            /\ referenceLock' = "callback" /\ callback' = 1
                            /\ UNCHANGED <<mapLock, cleanup>>
CallbackAcquireMap == /\ callback = 1 /\ mapLock = "free"
                      /\ mapLock' = "callback" /\ callback' = 2
                      /\ UNCHANGED <<referenceLock, cleanup>>
CallbackReturn == /\ callback = 2
                  /\ referenceLock' = "free" /\ mapLock' = "free" /\ callback' = 3
                  /\ UNCHANGED cleanup
CleanupAcquireMap == /\ cleanup = 0 /\ mapLock = "free"
                     /\ mapLock' = "cleanup" /\ cleanup' = 1
                     /\ UNCHANGED <<referenceLock, callback>>
CleanupDetach == /\ cleanup = 1
                 /\ mapLock' = IF DetachBeforeRelease THEN "free" ELSE "cleanup"
                 /\ cleanup' = 2
                 /\ UNCHANGED <<referenceLock, callback>>
CleanupAcquireReference == /\ cleanup = 2 /\ referenceLock = "free"
                           /\ referenceLock' = "cleanup" /\ cleanup' = 3
                           /\ UNCHANGED <<mapLock, callback>>
CleanupReturn == /\ cleanup = 3
                 /\ referenceLock' = "free" /\ mapLock' = "free" /\ cleanup' = 4
                 /\ UNCHANGED callback
CallbackStep == CallbackAcquireReference \/ CallbackAcquireMap \/ CallbackReturn
CleanupStep == CleanupAcquireMap \/ CleanupDetach \/ CleanupAcquireReference \/ CleanupReturn
Next == CallbackStep \/ CleanupStep \/ UNCHANGED vars
Spec == Init /\ [][Next]_vars /\ WF_vars(CallbackStep) /\ WF_vars(CleanupStep)
TypeOK == /\ referenceLock \in {"free", "callback", "cleanup"}
          /\ mapLock \in {"free", "callback", "cleanup"}
          /\ callback \in 0..3 /\ cleanup \in 0..4
NoLockCycle == ~(callback = 1 /\ cleanup = 2 /\ referenceLock = "callback" /\ mapLock = "cleanup")
Retires == <>(callback = 3 /\ cleanup = 4)
====================================================================
