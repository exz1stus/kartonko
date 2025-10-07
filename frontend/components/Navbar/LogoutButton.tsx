import { useAuth } from '@/app/contexts/AuthContext';
import React from 'react'
import Button from '../template/Button';

const LogoutButton: React.FC = () => {
    const { logout } = useAuth();

    return (
        <div>
            <Button text={"Log out"} onClick={async () => await logout()} />
        </div>
    )
}

export default LogoutButton
