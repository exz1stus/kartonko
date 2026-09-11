import { cn } from "@/lib/utils";
import { Turnstile, TurnstileInstance } from "@marsidev/react-turnstile";
import React, { useEffect, useRef, useState } from "react";

interface Props extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    onVerifySuccess: (token: string) => void;
    verifyingText?: string;
    resetKey?: number;
}

const CaptchaButton = ({
    children,
    verifyingText = "Verifying...",
    className,
    type = "button",
    disabled,
    onVerifySuccess,
    resetKey,
    ...props
}: Props) => {
    const [token, setToken] = useState<string | null>(null);
    const turnstileRef = useRef<TurnstileInstance>(null);

    useEffect(() => {
        reset();
    }, [resetKey]);

    const reset = () => {
        setToken(null);
        turnstileRef.current?.reset();
    };
    const handleReset = () => {
        reset();
    };
    const handleSuccess = (newToken: string) => {
        setToken(newToken);
        onVerifySuccess(newToken);
    };
    const handleError = () => {
        reset();
    };
    const handleExpire = () => {
        reset();
    };

    return (
        <div className="flex flex-col items-center gap-4">
            <Turnstile
                ref={turnstileRef}
                className={`${token && "hidden"} flex justify-center`}
                siteKey={process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY!}
                onSuccess={handleSuccess}
                onExpire={handleExpire}
                onError={handleError}
                onReset={handleReset}
            />
            {token && (
                <button
                    {...props}
                    className={cn(
                        "px-12 py-1 border border-surface-20 hover:border-surface-30 rounded-2xl w-96 transition cursor-pointer glass",
                        className,
                    )}
                    type={type}
                    disabled={disabled || !token}
                >
                    {children}
                </button>
            )}
        </div>
    );
};

export default CaptchaButton;
