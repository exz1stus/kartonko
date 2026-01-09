import React from "react";
import FancySpan from "../template/FancySpan";

interface UsernameButtonProps {
    username: string;
}

const UsernameButton: React.FC<UsernameButtonProps> = ({ username }: UsernameButtonProps) => {
    const clickUsername = () => {
        window.location.href = `/user/${username}`;
    };

    return (
        <div className="font-bold text-lg hover:cursor-pointer" onClick={clickUsername}>
            <FancySpan word={username} />
        </div>
    );
};

export default UsernameButton;
