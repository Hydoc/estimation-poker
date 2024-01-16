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
  guess: number;
  role: "developer";
};

export type UserOverview = {
  productOwnerList: ProductOwner[];
  developerList: Developer[];
};
