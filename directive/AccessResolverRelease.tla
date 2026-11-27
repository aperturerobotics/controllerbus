------------------------- MODULE AccessResolverRelease -------------------------
EXTENDS Naturals, TLC
\* Each getter acquisition has one release owner. Removed-value callbacks only
\* notify that owner, including callbacks queued before unregistration and run
\* after a later acquisition starts. Cancellation also notifies the same owner.
DirectReleaseOnRemoval == FALSE
SharedNotification == FALSE
Acquisitions == 1..2
VARIABLES current, phase, notified, ready, removed, queued, releases, stopped
vars == <<current, phase, notified, ready, removed, queued, releases, stopped>>
Init == /\ current = 1 /\ phase = "acquired" /\ notified = {}
        /\ ready = {} /\ removed = {} /\ queued = {} /\ releases = [i \in Acquisitions |-> 0]
        /\ stopped = FALSE
Publish == /\ phase = "acquired" /\ phase' = "published"
           /\ UNCHANGED <<current, notified, ready, removed, queued, releases, stopped>>
Remove == /\ phase = "published" /\ current \notin removed
          /\ removed' = removed \cup {current} /\ queued' = queued \cup {current}
          /\ UNCHANGED <<current, phase, notified, ready, releases, stopped>>
Callback(i) == /\ i \in queued /\ queued' = queued \ {i}
               /\ notified' = notified \cup {i}
               /\ ready' = ready \cup {IF SharedNotification THEN current ELSE i}
               /\ releases' = IF DirectReleaseOnRemoval
                              THEN [releases EXCEPT ![i] = @ + 1] ELSE releases
               /\ UNCHANGED <<current, phase, removed, stopped>>
Invalidate == /\ phase = "published" /\ current \notin notified
              /\ notified' = notified \cup {current}
              /\ ready' = ready \cup {current}
              /\ UNCHANGED <<current, phase, removed, queued, releases, stopped>>
Cancel == /\ ~stopped /\ stopped' = TRUE
          /\ notified' = notified \cup {current}
          /\ ready' = ready \cup {current}
          /\ UNCHANGED <<current, phase, removed, queued, releases>>
Release == /\ phase \in {"acquired", "published"} /\ current \in ready
           /\ releases' = [releases EXCEPT ![current] = @ + 1] /\ phase' = "released"
           /\ UNCHANGED <<current, notified, ready, removed, queued, stopped>>
Reject == /\ phase = "acquired" /\ phase' = "released"
          /\ releases' = [releases EXCEPT ![current] = @ + 1]
          /\ notified' = notified \cup {current} /\ ready' = ready \cup {current}
          /\ UNCHANGED <<current, removed, queued, stopped>>
Replace == /\ phase = "released" /\ current < 2 /\ ~stopped
           /\ current' = current + 1 /\ phase' = "acquired"
           /\ UNCHANGED <<notified, ready, removed, queued, releases, stopped>>
Next == Publish \/ Remove \/ (\E i \in Acquisitions: Callback(i)) \/ Invalidate
        \/ Cancel \/ Release \/ Reject \/ Replace
Spec == Init /\ [][Next]_vars /\ WF_vars(Release)
AtMostOnce == \A i \in Acquisitions: releases[i] <= 1
NoCrossAcquisitionRelease == \A i \in Acquisitions: releases[i] > 0 => i \in notified
NotifiedEventuallyReleased == \A i \in Acquisitions: (i \in notified) ~> (releases[i] = 1)
===============================================================================
