import { Role, RoomState, RoundState, UserOverview } from "../../src/types/room";
import { just, nothing } from "@kaumlaut/pure/maybe";
import { Permissions } from "../../src/types/room";
import { succeed } from "@kaumlaut/pure/fetch-state";

export class RoomStateBuilder {
  private constructor(private roomState: RoomState) {}

  withShowAllGuesses(showAllGuesses: boolean): RoomStateBuilder {
    return new RoomStateBuilder({
      ...this.roomState,
      showAllGuesses,
    });
  }

  withIssueToGuess(issue: string): RoomStateBuilder {
    return new RoomStateBuilder({
      ...this.roomState,
      issueToGuess: just(issue),
    });
  }

  withRoundState(state: RoundState): RoomStateBuilder {
    return new RoomStateBuilder({
      ...this.roomState,
      roundState: state,
    });
  }

  withRoomIsLocked(isLocked: boolean): RoomStateBuilder {
    return new RoomStateBuilder({
      ...this.roomState,
      roomIsLocked: isLocked,
    });
  }

  withConnected(connected: boolean): RoomStateBuilder {
    return new RoomStateBuilder({
      ...this.roomState,
      isConnected: connected,
    });
  }

  withUsers(users: UserOverview): RoomStateBuilder {
    return new RoomStateBuilder({
      ...this.roomState,
      users: succeed(users),
    });
  }

  withPermissions(permissions: Permissions): RoomStateBuilder {
    return new RoomStateBuilder({
      ...this.roomState,
      permissions: permissions,
    });
  }

  withRole(role: Role): RoomStateBuilder {
    return new RoomStateBuilder({
      ...this.roomState,
      role: just(role),
    });
  }

  withName(name: string): RoomStateBuilder {
    return new RoomStateBuilder({
      ...this.roomState,
      name: just(name),
    });
  }

  withId(id: string): RoomStateBuilder {
    return new RoomStateBuilder({
      ...this.roomState,
      id: just(id),
    });
  }

  static init(): RoomStateBuilder {
    return new RoomStateBuilder({
      issues: [],
      id: just("a-room-id"),
      roundInProgress: false,
      name: "Tester",
      isConnected: true,
      roomIsLocked: false,
      users: succeed([
        { name: "Test", isDone: false, role: Role.Developer },
        { name: "Product Owner Test", role: Role.ProductOwner },
      ]),
      possibleGuesses: [
        { guess: 1, description: "Bis zu 4 Std." },
        { guess: 2, description: "Bis zu 8 Std." },
        { guess: 3, description: "Bis zu 3 Tagen" },
        { guess: 4, description: "Bis zu 5 Tagen" },
        { guess: 5, description: "Mehr als 5 Tage" },
      ],
      role: just(Role.Developer),
      doSkip: false,
      issueToGuess: nothing(),
      showAllGuesses: false,
      developerDone: [],
      guess: nothing(),
      roundState: RoundState.Waiting,
      permissions: {
        canLockRoom: false,
        key: "",
      },
    });
  }

  build(): RoomState {
    return this.roomState;
  }
}
