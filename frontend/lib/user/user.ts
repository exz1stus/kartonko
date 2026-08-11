import { UserData } from "./dto";

export function isModerator(user: UserData) {
    return user.privilege === "Moderator";
}
