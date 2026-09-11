"use server";
import { isModerator } from "@/lib/user/user";
import { forbidden, notFound } from "next/navigation";
import Image from "next/image";
import GalleryServer from "@/components/Gallery/GalleryServer";
import TimeField from "@/components/TimeField";
import EditUser from "@/components/EditUser";
import { getUserName } from "@/lib/api/generated/server";
import { UserDataResponse } from "@/lib/api/generated/model";
import { getLoggedUser } from "@/lib/user/user.server";

const UserPage = async ({
    params,
}: {
    params: Promise<{ username: string }>;
}) => {
    const { username } = await params;

    let user = await getUserName(username);
    let loggedUser = await getLoggedUser();

    let hasEditPermission =
        loggedUser !== null &&
        (isModerator(loggedUser) || loggedUser.id === user.id);

    const pictureURL: string =
        user.picture_url?.replace("s96-c", "s256-c") ||
        "https://cdn-icons-png.flaticon.com/512/149/149071.png";

    return (
        // <div className="flex justify-center w-full h-full">
        //     <div className="flex flex-col w-full h-full">
        //         <div className="flex justify-between min-h-[10vh]">
        //             <Image
        //                 src={pictureURL}
        //                 alt="User picture"
        //                 width={256}
        //                 height={256}
        //             />
        //             <div>
        //                 {" "}
        //                 <span className="text-xl">{user.username}</span>
        //                 <div className="flex justify-between">
        //                     <span>Privilege: </span>
        //                     <span>{user.privilege}</span>
        //                 </div>
        //                 <div className="flex justify-between">
        //                     <span>Joined at: </span>
        //                     <span>{user.joined_at}</span>
        //                 </div>
        //                 <div className="flex justify-between">
        //                     <span>Last seen: </span>
        //                     <TimeField time={user.last_seen} />
        //                 </div>
        //                 <EditUser
        //                     user={user}
        //                     hasPermission={hasEditPermission}
        //                 />
        //             </div>
        //         </div>
        //         <div className="flex-3 min-w-0 h-full">
        //             <GalleryServer
        //                 initialFetchSize={50}
        //                 initialQuery={{ userID: user.id }}
        //             />
        //         </div>
        //     </div>
        // </div>
        <div className="flex flex-row justify-center w-full h-full">
            <div className="flex flex-row border-r w-full xl:w-2/3 h-full">
                <div className="flex flex-col flex-1 bg-surface-0/80 border-x min-w-0 h-full glass">
                    <Image
                        src={pictureURL}
                        className="w-full"
                        alt="User picture"
                        width={256}
                        height={256}
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
