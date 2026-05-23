import { isModerator, UserData } from "@/lib/user/user";
import { notFound } from "next/navigation";
import Image from "next/image";
import GalleryServer from "@/components/Gallery/GalleryServer";
import TimeField from "@/components/TimeField";
import { serverFetch } from "@/lib/serverFetch";
import EditUser from "@/components/EditUser";
import { getLoggedUserServer } from "@/lib/user/user.server";

const UserPage = async ({
    params,
}: {
    params: Promise<{ username: string }>;
}) => {
    const { username } = await params;

    const fetchUser = async (username: string): Promise<UserData> => {
        try {
            const res = await serverFetch(`/user/${username}`);

            if (!res.ok) return notFound();

            const user = await res.json();

            if (!user) return notFound();

            return user;
        } catch (err) {
            console.error("Failed to fetch user:", err);
            return notFound();
        }
    };

    let user = await fetchUser(username);
    let loggedUser = await getLoggedUserServer();

    let hasEditPermission =
        loggedUser !== null &&
        (isModerator(loggedUser) || loggedUser.id === user.id);

    const pictureURL: string =
        user.picture_url?.replace("s96-c", "s256-c") ||
        "https://cdn-icons-png.flaticon.com/512/149/149071.png";

    return (
        <div className="flex flex-row justify-center w-full h-full">
            <div className="flex flex-row border-r w-full xl:w-2/3 h-full">
                <div className="flex flex-col flex-1 bg-surface-0/80 border-x h-full glass">
                    <Image
                        src={pictureURL}
                        className="w-full"
                        alt="User picture"
                        width={256}
                        height={256}
                    />
                    <div className="flex flex-col px-1 py-2">
                        <span className="text-5xl">{user.username}</span>
                        <div className="flex justify-between">
                            <span>Privileage: </span>
                            <span>{user.privileage}</span>
                        </div>
                        <div className="flex justify-between">
                            <span>Joined at: </span>
                            <span>{user.joined_at}</span>
                        </div>
                        <div className="flex justify-between">
                            <span>Last seen: </span>
                            <TimeField time={user.last_seen} />
                        </div>
                        <EditUser
                            user={user}
                            hasPermission={hasEditPermission}
                        />
                    </div>
                </div>
                <div className="flex-5 w-full">
                    <GalleryServer
                        initialFetchSize={50}
                        initialQuery={{ userID: user.id }}
                    />
                </div>
            </div>
        </div>
    );
};

export default UserPage;
