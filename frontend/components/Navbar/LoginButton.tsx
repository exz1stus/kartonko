"use client";
import { useRouter } from "next/navigation";

const LoginButton: React.FC = () => {
    const router = useRouter();
    return (
        <button
            onClick={() =>
                router.push(
                    `/login?redirect=${encodeURIComponent(window.location.pathname || "/")}`,
                )
            }
            className="w-auto h-10 cursor-pointer"
        >
            <span>Log in</span>
        </button>
    );
};

export default LoginButton;
