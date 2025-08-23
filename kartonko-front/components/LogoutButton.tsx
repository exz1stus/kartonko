import { useAuth } from '@/app/AuthContext';
import React from 'react'

const LogoutButton: React.FC = () => {
    const { logout } = useAuth();

    return (
        <div>
            <button onClick={async () => await logout()}>Log out</button>
        </div>
    )
}

export default LogoutButton
