import Link from "next/link";
import React from "react";

const Logo = () => {
    return (
        <Link
            href="/"
            className="left-1/2 absolute -translate-x-1/2 cursor-pointer"
        >
            <span className="font-bold text-2xl">kartonko</span>
        </Link>
    );
};

export default Logo;
