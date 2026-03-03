export interface UserData {
    id: number;
    username: string;
    privileage: string;
    picture_url: string;
    joined_at: string;
    last_seen: string;
}

export function isModerator(user: UserData) {
    return user.privileage === "Moderator";
}
