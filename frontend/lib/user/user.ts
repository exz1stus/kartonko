import { UserDataResponse } from "../api/generated/model";

export function isModerator(user: Pick<UserDataResponse, "privilege">) {
    return user.privilege === "Moderator";
}
