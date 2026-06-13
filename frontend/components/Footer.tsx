import Link from "next/link";
import React from "react";

const Footer = () => {
    return (
        <div className="flex flex-row justify-end gap-4 bg-surface-0/80 px-4 w-full glass">
            <span className="inline-flex">
                <Link href="/about">about</Link>
            </span>
            <span className="inline-flex gap-1">
                <Link href="https://github.com/exz1stus/kartonko">github</Link>
            </span>
        </div>
    );
};

export default Footer;
