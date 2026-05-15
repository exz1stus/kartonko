import React from "react";
import Image from "next/image";
import "./globals.css";

const NotFound = () => {
    return (
        <div className="flex justify-center items-center gap-20 bg-surface-10 w-full min-h-screen">
            <Image
                src="/images/megusta.png"
                className="max-w-[50vw] animate-bounce"
                alt="megusta"
                width={256}
                height={256}
            />
            <span className="text-5xl">Отакої 404</span>
        </div>
    );
};

export default NotFound;
