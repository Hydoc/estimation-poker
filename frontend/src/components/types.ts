export enum Role {
  ProductOwner = "product-owner",
  Developer = "developer",
  Empty = "",
}

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

export enum RoundState {
  Waiting,
  InProgress,
  End,
}

export type PossibleGuess = {
  guess: number;
  description: string;
};

export type Permissions = {
  room: {
    canLock: boolean;
    key?: string;
  };
};

export type FetchActiveRoomsResponse = {
    rooms: ActiveRoom[] | null;
};

export type ActiveRoom = {
    id: string;
    playerCount: number;
}