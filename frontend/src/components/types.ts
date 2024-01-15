export enum Role {
  ProductOwner = "product-owner",
  Developer = "developer",
  Empty = "",
}

export type ProductOwner = {
  name: string;
  role: "product-owner";
};

export type Developer = {
  guess: number;
  name: string;
  role: "developer";
};
