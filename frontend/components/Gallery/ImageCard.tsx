import PerspectiveCard from "./PerspectiveCard";
import ImageCardServer from "./ImageCardServer";
import React, { useRef } from "react";
import { useRouter } from "next/navigation";
import ImageMetadata from "@/lib/image";

interface Props {
    image: ImageMetadata;
    className?: string;
    style?: React.CSSProperties;
}

const ImageCard: React.FC<Props> = ({ image, className, style }) => {
    const selfRef = useRef<HTMLDivElement>(null);
    const router = useRouter();
    const onClick = () => {
        router.push(`/image/${image.filename}`);
    };

    const onLoad = () => {
        selfRef.current?.classList.remove("opacity-0");
        selfRef.current?.clientHeight;
    };

    return (
        <div ref={selfRef} className={className + " opacity-0 "} style={style} onClick={onClick}>
            <PerspectiveCard>
                <ImageCardServer image={image} onLoad={onLoad} />
            </PerspectiveCard>
        </div>
    );
};

export default ImageCard;
