# Sequence

## Admin

```
Admin -> Server: Create Participant
Server -> Admin: Participant with conference and waiting room token
Admin -> LiveKit: Join conference room with token
Admin -> LiveKit: Join waitiing room with token (to admit others)
Admin is in the conference!
```

## Guest

```
Participant -> Server: Create Participant
Server -> Participant: Participant with waiting room token
Participant -> LiveKit: Join waiting room with token
Participant waits...

[alt: Deny Participant]
~~~~~~~~~~~~~~~~~~~~~~~
LiveKit -> Admin: Participant connected
Admin -> Server: Update Participant with status=denied
Server -> LiveKit: Send data to waiting room "Participant Denied"
LiveKit -> Participant: Participant Denied
Participant leaves!

[alt: Admit Participant]
~~~~~~~~~~~~~~~~~~~~~~~
LiveKit -> Admin: Participant connected
Admin -> Server: Update Participant with status=admitted
Server -> LiveKit: Send data to waiting room "Participant Admitted"
LiveKit -> Participant: Participant Admitted
Participant -> Server: Retrieve Participant
Server -> Participant: Participant with conference room token
Participant -> LiveKit: Join conference room with token
Participant is in the conference!
```
