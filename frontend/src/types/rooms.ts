import {
    isListOf,
    isNumber,
    isObjectWithKeysMatchingGuard,
    isString
} from "@kaumlaut/pure/error-aware-guard";

type ActiveRoom = {
    id: string;
    playerCount: number;
};

export type ActiveRooms = {
    rooms: ActiveRoom[];
}

const isActiveRoom = isObjectWithKeysMatchingGuard<ActiveRoom>({
    id: isString,
    playerCount: isNumber,
});

export const isActiveRooms = isObjectWithKeysMatchingGuard<ActiveRooms>({
    rooms: isListOf(isActiveRoom),
});

export const isRoomCreated = isObjectWithKeysMatchingGuard<{ id: string }>({
    id: isString,
});
