import { UserDataResponse } from "@/lib/api/generated/model";
import ProfileImage from "@/components/ProfileImage";

interface Props {
    user: Pick<UserDataResponse, "picture_url">;
}

const UserPicture = ({ user }: Props) => {
    return (
        <div className="rounded-full w-10 h-10 cursor-pointer">
            <ProfileImage
                src={user.picture_url}
                size={40}
                className="rounded-full w-full h-full object-cover"
            />
        </div>
    );
};

export default UserPicture;
