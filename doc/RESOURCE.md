# Resource

## User

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

## Auth

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

type AuthorizationAccessTokenPayload = {
  id: string;
  userId: string;
};
```

## Session

- `/session` _GET_

```ts
type Session = {
  user: User | null;
};
```

## Meeting

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

type ParticipantWithToken = Participant & {
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
```
