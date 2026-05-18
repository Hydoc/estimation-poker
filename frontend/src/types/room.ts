import type { Maybe } from "@kaumlaut/pure/maybe";
import {
  isAlways,
  isBool,
  isExactString,
  isFalse,
  isListOf,
  isNonEmptyString,
  isNull,
  isNullOr,
  isNumber,
  isObjectWithKeysMatchingGuard,
  isOneOf,
  isOneStringOf,
  isString,
} from "@kaumlaut/pure/error-aware-guard";
import type { FetchState } from "@kaumlaut/pure/fetch-state";

export type User = {
  name: string;
};

export type ProductOwner = User & {
  role: Role.ProductOwner;
};

export type Developer = User & {
  isDone: boolean;
  role: Role.Developer;
};

export type DeveloperDone = {
  doSkip: boolean;
  guess: number;
  name: string;
  role: Role.Developer;
};

export type UserOverview = (ProductOwner | Developer)[];

export enum Role {
  ProductOwner = "product-owner",
  Developer = "developer",
  Empty = "",
}

export type ConnectionState = {
  canConnect: boolean;
  reason: "wrong password" | "round already started" | "username already taken" | "";
};

export type Issue = {
  title: string;
  guess: number;
};

export type RoomMetadata = {
  exists: boolean;
  isLocked: boolean;
};

export type Permissions = {
  canLockRoom: boolean;
  key: string;
};

export enum RoundState {
  Waiting,
  InProgress,
  End,
}

export type RoomState = Readonly<{
  id: Maybe<string>;
  guess: Maybe<number>;
  role: Maybe<Role>;
  name: Maybe<string>;
  doSkip: boolean;
  issueToGuess: Maybe<string>;
  roundState: RoundState;
  users: Maybe<UserOverview>;
  showAllGuesses: boolean;
  roomIsLocked: boolean;
  roundInProgress: boolean;
  developerDone: DeveloperDone[];
  issues: any[];
  isConnected: boolean;
  permissions: Permissions;
  possibleGuesses: PossibleGuess[];
}>;

export type SendableWebsocketMessageType =
  | "estimate"
  | "guess"
  | "reveal"
  | "new-round"
  | "lock-room"
  | "skip"
  | "open-room"
  | "add-issue";

export type SendableWebsocketMessage = {
  type: SendableWebsocketMessageType;
  data?: any;
};

export type ReceivableWebsocketMessage = {
  type:
    | "leave"
    | "estimate"
    | "reveal"
    | "developer-guessed"
    | "everyone-done"
    | "you-guessed"
    | "you-skipped"
    | "new-round"
    | "room-locked"
    | "developer-skipped"
    | "room-opened"
    | "issues"
    | "permissions"
    | "users";
  data?: any;
};

export type PossibleGuess = {
  guess: number;
  description: string;
};

const isProductOwner = isObjectWithKeysMatchingGuard<ProductOwner>({
  name: isNonEmptyString,
  role: isExactString(Role.ProductOwner),
});

const isDeveloper = isObjectWithKeysMatchingGuard<Developer>({
  name: isNonEmptyString,
  isDone: isBool,
  role: isExactString(Role.Developer),
});

const isDeveloperDone = isObjectWithKeysMatchingGuard<DeveloperDone>({
  doSkip: isBool,
  guess: isNumber,
  name: isString,
  role: isExactString(Role.Developer),
});

export const isUserOverview = isListOf<ProductOwner | Developer>(
  isOneOf(isDeveloper, isProductOwner),
);

export const isReceivableWebsocketMessage =
  isObjectWithKeysMatchingGuard<ReceivableWebsocketMessage>({
    type: isOneStringOf([
      "leave",
      "estimate",
      "reveal",
      "developer-guessed",
      "everyone-done",
      "you-guessed",
      "you-skipped",
      "new-round",
      "room-locked",
      "developer-skipped",
      "room-opened",
      "issues",
      "permissions",
      "users",
    ]),
    data: isAlways,
  });

export const isLeaveWebsocketMessage = isObjectWithKeysMatchingGuard<{
  type: "leave";
  data: string;
}>({
  type: isExactString("leave"),
  data: isString,
});

export const isUsersWebsocketMessage = isObjectWithKeysMatchingGuard<{
  type: "users";
  data: (ProductOwner | Developer)[];
}>({
  type: isExactString("users"),
  data: isListOf(isOneOf(isDeveloper, isProductOwner)),
});

export const isEstimateWebsocketMessage = isObjectWithKeysMatchingGuard<{
  type: "estimate";
  data: string;
}>({
  type: isExactString("estimate"),
  data: isString,
});

export const isYouGuessedWebsocketMessage = isObjectWithKeysMatchingGuard<{
  type: "you-guessed";
  data: number;
}>({
  type: isExactString("you-guessed"),
  data: isNumber,
});

export const isYouSkippedWebsocketMessage = isObjectWithKeysMatchingGuard<{
  type: "you-skipped";
  data: null;
}>({
  type: isExactString("you-skipped"),
  data: isNull,
});

export const isEveryoneDoneWebsocketMessage = isObjectWithKeysMatchingGuard<{
  type: "everyone-done";
  data: null;
}>({
  type: isExactString("everyone-done"),
  data: isNull,
});

export const isRevealWebsocketMessage = isObjectWithKeysMatchingGuard<{
  type: "reveal";
  data: DeveloperDone[];
}>({
  type: isExactString("reveal"),
  data: isListOf(isDeveloperDone),
});

export const isRoomLockedWebsocketMessage = isObjectWithKeysMatchingGuard<{
  type: "room-locked";
  data: null;
}>({
  type: isExactString("room-locked"),
  data: isNull,
});

export const isRoomOpenedWebsocketMessage = isObjectWithKeysMatchingGuard<{
  type: "room-opened";
  data: null;
}>({
  type: isExactString("room-opened"),
  data: isNull,
});

export const isNewRoundWebsocketMessage = isObjectWithKeysMatchingGuard<{
  type: "new-round";
  data: null;
}>({
  type: isExactString("new-round"),
  data: isNull,
});

export const isIssuesWebsocketMessage = isObjectWithKeysMatchingGuard<{
  type: "issues";
  data: null;
}>({
  type: isExactString("issues"),
  data: isNull,
});

export const isPermissionsWebsocketMessage = isObjectWithKeysMatchingGuard<{
  type: "permissions";
  data: Permissions;
}>({
  type: isExactString("permissions"),
  data: isObjectWithKeysMatchingGuard<Permissions>({
    key: isString,
    canLockRoom: isBool,
  }),
});

const isPossibleGuess = isObjectWithKeysMatchingGuard<PossibleGuess>({
  description: isNonEmptyString,
  guess: isNumber,
});

const isIssue = isObjectWithKeysMatchingGuard<Issue>({
  title: isNonEmptyString,
  guess: isNumber,
});

export const isIssues = isListOf(isIssue);

export const isRoomStateResponse = isObjectWithKeysMatchingGuard<{
  isLocked: boolean;
  inProgress: boolean;
  issues: Issue[];
  possibleGuesses: PossibleGuess[];
}>({
  isLocked: isBool,
  inProgress: isBool,
  issues: isListOf(isIssue),
  possibleGuesses: isListOf(isPossibleGuess),
});

export const isRoomMetadata = isObjectWithKeysMatchingGuard<RoomMetadata>({
  exists: isBool,
  isLocked: isBool,
});

export const isConnectionState = isObjectWithKeysMatchingGuard<ConnectionState>({
  canConnect: isBool,
  reason: isOneStringOf(["round already started", "username already taken", "wrong password", ""]),
});

export const isWrongPasswordConnectionStatus = isObjectWithKeysMatchingGuard<ConnectionState>({
  canConnect: isFalse,
  reason: isExactString("wrong password"),
});

export const isRoundAlreadyStartedConnectionStatus = isObjectWithKeysMatchingGuard<ConnectionState>(
  {
    canConnect: isFalse,
    reason: isExactString("round already started"),
  },
);

export const isUsernameAlreadyTakenConnectionStatus =
  isObjectWithKeysMatchingGuard<ConnectionState>({
    canConnect: isFalse,
    reason: isExactString("username already taken"),
  });
