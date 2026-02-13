export interface UserData {
    username: string;
    privileage: string;
    picture_url: string;
    joined_at: string;
    last_seen: string;
}

const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

const getUserById = async (id: number): Promise<UserData | null> => {
    try {
        const response = await fetch(`${API_ORIGIN}/user?id=${id}`);

        if (!response.ok) {
            if (response.status === 404) {
                return null;
            }
            throw new Error(`Failed to fetch user: ${response.statusText}`);
        }

        const userData: UserData = await response.json();
        return userData;
    } catch (error) {
        console.error("Error fetching user:", error);
        return null;
    }
};

export { getUserById };
