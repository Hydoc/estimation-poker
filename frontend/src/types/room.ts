import type { Maybe } from "@kaumlaut/pure/maybe";
import {
  isBool,
  isExactString,
  isFalse,
  isListOf,
  isNonEmptyString,
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
  users: FetchState<UserOverview>;
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
    | "join"
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
    | "permissions";
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

export const isUserOverview = isListOf<ProductOwner | Developer>(
  isOneOf(isDeveloper, isProductOwner),
);

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

export const isPermissions = isObjectWithKeysMatchingGuard<Permissions>({
  canLockRoom: isBool,
  key: isString,
});
