# Sequence

## Admin

```
Admin --> Server: Create Participant
Admin <<- Server: Participant with conference and waiting room token
Admin --> LiveKit: Join conference room with token
Admin --> LiveKit: Join waitiing room with token (to admit others)
Admin is in the conference!
```

## Guest

```
Participant --> Server: Create Participant
Participant <<- Server: Participant with waiting room token
Participant --> LiveKit: Join waiting room with token
Participant waits...

[alt: Deny Participant]
~~~~~~~~~~~~~~~~~~~~~~~
Admin <== LiveKit: Participant connected
Admin --> Server: Update Participant with status=denied
Server ~~> LiveKit: Send data to waiting room "Participant Denied"
Participant <== LiveKit: Participant Denied
Participant leaves!

[alt: Admit Participant]
~~~~~~~~~~~~~~~~~~~~~~~
Admin <== LiveKit: Participant connected
Admin --> Server: Update Participant with status=admitted
Server ~~> LiveKit: Send data to waiting room "Participant Admitted"
Participant <== LiveKit: Participant Admitted
Participant --> Server: Retrieve Participant
Participant <<- Server: Participant with conference room token
Participant --> LiveKit: Join conference room with token
Participant is in the conference!
```
