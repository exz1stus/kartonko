export interface UserData {
    id: number;
    username: string;
    privilege: string;
    picture_url: string;
    joined_at: string;
    last_seen: string;
}

export function isModerator(user: UserData) {
    return user.privilege === "Moderator";
}
