import type { Maybe } from "@kaumlaut/pure/maybe";
import {
  isBool,
  isExactString,
  isListOf,
  isNonEmptyString,
  isNumber,
  isObjectWithKeysMatchingGuard,
  isOneOf,
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

export type Permissions = {
  room: {
    canLock: boolean;
    key?: string;
  };
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
    | "issues";
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

export const isPossibleGuesses = isListOf(isPossibleGuess);

export const isRoomStateResponse = isObjectWithKeysMatchingGuard<{
  isLocked: boolean;
  inProgress: boolean;
}>({
  isLocked: isBool,
  inProgress: isBool,
});

export const isPermissions = isObjectWithKeysMatchingGuard({
  permissions: isObjectWithKeysMatchingGuard({
    room: isOneOf(
      isObjectWithKeysMatchingGuard({
        canLock: isBool,
      }),
      isObjectWithKeysMatchingGuard({
        canLock: isBool,
        key: isNonEmptyString,
      }),
    ),
  }),
});
