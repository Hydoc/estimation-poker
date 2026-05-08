import type {Maybe} from "@kaumlaut/pure/maybe";
import type {DeveloperDone, RoundState, UserOverview} from "@/components/types.ts";

export type RoomState = Readonly<{
  id: Maybe<string>;
  guess: Maybe<number>;
  doSkip: boolean;
  issueToGuess: Maybe<string>;
  roundState: RoundState;
  users: UserOverview;
  notifications: string[]
  showAllGuesses: boolean;
  roomIsLocked: boolean;
  developerDone: DeveloperDone[];
  issues: any[];
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
