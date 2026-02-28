import { getUserByIdServer, isModerator } from "@/lib/user";
import UserPicture from "@/components/Navbar/UserPicture";
import Link from "next/link";

interface Props {
    id: number;
}

const UserElement = async ({ id }: Props) => {
    const user = await getUserByIdServer(id);

    if (!user) return <div>Not found</div>;

    return (
        <Link className="flex flex-row items-center gap-2" href={`/user/${user.username}`}>
            <UserPicture user={user} />
            <div className={isModerator(user) ? "text-orange-400" : ""}>{user.username}</div>
        </Link>
    );
};

export default UserElement;
