import { isModerator, UserData } from "@/lib/user";
import UserPicture from "@/components/Navbar/UserPicture";
import Link from "next/link";

const UserElement = ({ user }: { user: UserData | null }) => {
    if (!user) return <div>Not found</div>;

    return (
        <Link
            className="flex flex-row items-center gap-2"
            href={`/user/${user.username}`}
        >
            <UserPicture user={user} />
            <div className={isModerator(user) ? "text-orange-400" : ""}>
                {user ? user.username : "not found"}
            </div>
        </Link>
    );
};

export default UserElement;
