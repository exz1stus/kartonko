import { UserData } from "@/lib/user";

interface Props {
    user: UserData;
}

const UserPicture = ({ user }: Props) => {
    const url: string =
        user?.picture_url || "https://cdn-icons-png.flaticon.com/512/149/149071.png";

    return (
        <div className="rounded-full w-10 h-10 cursor-pointer">
            <img src={url} className="rounded-full w-full h-full" alt="User picture" />
        </div>
    );
};

export default UserPicture;
