"use server";
import { isModerator } from "@/lib/user/user";
import ProfileImage from "@/components/ProfileImage";
import GalleryServer from "@/components/Gallery/GalleryServer";
import TimeField from "@/components/TimeField";
import EditUser from "@/components/EditUser";
import { getUserByName } from "@/lib/api/generated/server";
import { getLoggedUser } from "@/lib/user/user.server";

const UserPage = async ({
    params,
}: {
    params: Promise<{ username: string }>;
}) => {
    const { username } = await params;

    let user = await getUserByName(username);
    let loggedUser = await getLoggedUser();

    let hasEditPermission =
        loggedUser !== null &&
        (isModerator(loggedUser) || loggedUser.id === user.id);

    const pictureURL = user.picture_url?.replace("s96-c", "s256-c");

    return (
        <div className="flex flex-row justify-center w-full h-full">
            <div className="flex flex-row border-r w-full xl:w-2/3 h-full">
                <div className="flex flex-col flex-1 bg-surface-0/80 border-x min-w-0 h-full glass">
                    <ProfileImage
                        src={pictureURL}
                        className="w-full"
                        size={256}
                    />
                    <div className="flex flex-col px-1 py-2">
                        <span className="text-xl">{user.username}</span>
                        <div className="flex justify-between">
                            <span>Privilege: </span>
                            <span>{user.privilege}</span>
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
                <div className="flex-3 min-w-0">
                    <GalleryServer
                        initialFetchSize={50}
                        initialQuery={{ user_id: user.id }}
                    />
                </div>
            </div>
        </div>
    );
};

export default UserPage;
