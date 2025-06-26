export enum Role {
  ProductOwner = "product-owner",
  Developer = "developer",
  Empty = "",
}

export type User = {
  name: string;
};

export type ProductOwner = User & {
  role: "product-owner";
};

export type Developer = User & {
  isDone: boolean;
  role: "developer";
};

export type DeveloperDone = {
  doSkip: boolean;
  guess: number;
  name: string;
  role: "developer";
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
