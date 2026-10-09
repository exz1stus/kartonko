import { UserDataResponse } from "../api/generated/model";

export function isModerator(user: UserDataResponse) {
    return user.privilege === "Moderator";
}
