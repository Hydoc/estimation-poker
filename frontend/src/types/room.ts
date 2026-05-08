import type { Maybe } from "@kaumlaut/pure/maybe";
import {
  type Developer,
  type DeveloperDone,
  type Permissions,
  type ProductOwner,
  Role,
  type RoundState,
  type UserOverview,
} from "@/components/types.ts";
import {
  isBool,
  isExactString,
  isListOf,
  isNonEmptyString,
  isObjectWithKeysMatchingGuard,
  isOneOf,
  isString,
  isUndefined,
} from "@kaumlaut/pure/error-aware-guard";
import type { FetchState } from "@kaumlaut/pure/fetch-state";

export type Round = Readonly<{
  state: RoundState;
  issueToGuess: string;
  users: Readonly<UserOverview>;
  developerDone: DeveloperDone[];
}>;

export type RoomState = Readonly<{
  id: Maybe<string>;
  guess: Maybe<number>;
  role: Maybe<Role>;
  name: Maybe<string>;
  doSkip: boolean;
  issueToGuess: Maybe<string>;
  roundState: RoundState;
  users: FetchState<UserOverview>;
  notifications: string[];
  showAllGuesses: boolean;
  roomIsLocked: boolean;
  roundInProgress: boolean;
  developerDone: DeveloperDone[];
  issues: any[];
  isConnected: boolean;
  permissions: Permissions;
}>;

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
