# LiveMeet Server

Web server for a simple video conferencing app written in golang with livekit.

## Core Features

- [x] Setup unit and integration tests
- [x] Sign in with Google
- [x] Users, auth & session
- [x] Meetings & participants
- [x] LiveKit integration
- [x] Github actions workflow
- [ ] Kubernetes manifests or helm chart

## Additional Features

- [ ] Admit guests into the meeting as guest
- [ ] Persist participants in meeting
- [ ] Cloud recording
- [ ] Auto generate API documentation?

## Develop

```bash
make all    # runs test, e2e and build
make build  # builds the binary
make clean  # cleans the binary
make test   # runs unit tests
make e2e    # runs integration tests
make run    # builds and runs binary
make dev    # runs the source and watches for changes
```
