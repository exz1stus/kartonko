import React from 'react'
import FancySpan from '../FancySpan';

interface UsernameButtonProps {
    username: string
}

const UsernameButton: React.FC<UsernameButtonProps> = ({ username }: UsernameButtonProps) => {
    const clickUsername = () => {
        window.location.href = `/user/${username}`;
    }

    return (
        <div className="text-lg font-bold hover:cursor-pointer" onClick={clickUsername}>
            <FancySpan word={username} />
        </div>
    )
}

export default UsernameButton
