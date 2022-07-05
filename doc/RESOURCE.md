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

- `/auth`
- `/auth/:authId`

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

## Meeting

- `/meetings`
- `/meetings?code=...`
- `/meetings/:meetingId`

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

- `/meetings/:meetingId/participants` _POST only_
- `/meetings/:meetingId/participants/:participantId` _For polling_

```ts
// expires when participant token created, admission denied or after 30m
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
