# Resource

## User

Defines an authorized user

```ts
type User = {
  id: string;
  name: string;
  imageUrl: string | null;
  provider: "google" | "facebook";
  providerResourceId: string;
  createdAt: string;
  updatedAt: string;
};
```

## Session

Defines the current authorized user

- `/session` _GET_

```ts
type Session = {
  user: User | null;
};
```

## Auth

Defines an authorization

- `/auth` _POST_
- `/auth/:authId` _PUT_

```ts
type Auth = {
  id: string;
  userId: string;
  refreshToken: string;
  refreshTokenExpiresAt: string;
  createdAt: string;
  updatedAt: string;
};

type AuthWithAccessToken = Auth & {
  scheme: "Bearer";
  accessToken: string;
  accessTokenExpiresAt: string;
};

type AuthAccessTokenPayload = {
  id: string;
  userId: string;
};

type AuthCreateBody = {
  googleCode: string;
};

type AuthRefreshBody = {
  refreshToken: string;
};
```

## Meeting

Defines a meeting

- `/meetings` _POST_
- `/meetings?code=...` _GET_
- `/meetings/:meetingId` _GET_

```ts
type Meeting = {
  id: string;
  userId: string;
  code: string;
  createdAt: string;
  updatedAt: string;
  expiresAt: string; // ttl: 365d
};
```

## Participant

Defines a participant in a meeting

- `/meetings/:meetingId/participants` _POST_
- `/meetings/:meetingId/participants/:participantId` _PUT_, _GET_ (poll)

```ts
// Participant expires when:
// - admission is granted and participant is retrieved
// - admission is denied and participant is retrieved
// - 30m has elapsed

type Participant = {
  id: string;
  meetingId: string;
  name: string;
  imageUrl: string | null;
  status: "waiting" | "admitted" | "denied";
  createdAt: string;
  updatedAt: string;
  expiresAt: string; // ttl: 30m
};

type ParticipantWithJoinToken = Participant & {
  token: string | null;
  tokenExpiresAt: string;
};

type ParticipantTokenPayload = {
  identity: string; // participantId
  metadata: string;
};

type ParticipantTokenMetadataPayload = {
  name: string;
  imageUrl: string | null;
};

type ParticipantCreateBody = {
  name?: string;
};

type ParticipantUpdateBody = {
  status: "admitted" | "denied";
};
```
