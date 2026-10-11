import { isModerator } from "@/lib/user/user";
import { UserDataResponse } from "@/lib/api/generated/model";
import UserPicture from "@/components/Navbar/UserPicture";
import Link from "next/link";

const UserElement = ({ user }: { user: Pick<UserDataResponse, "username" | "privilege" | "picture_url"> | null }) => {
    if (!user) return <div>Not found</div>;

    return (
        <Link
            className="flex flex-row items-center gap-2"
            href={`/user/${user.username}`}
        >
            <UserPicture user={user} />
            <div className={isModerator(user) ? "text-orange-400" : ""}>
                {user.username}
            </div>
        </Link>
    );
};

export default UserElement;
