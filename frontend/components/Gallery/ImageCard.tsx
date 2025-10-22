import { useEffect, useState } from 'react';
import PerspectiveCard from './PerspectiveCard';
import { noUse } from '@/app/AudioEffects';

interface Tag {
    Name: string
}

interface ImageData {
    filename: string
    tags: Tag[]
}

const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

const ImageCard: React.FC<ImageData> = ({ filename, tags }) => {
    const onMouseEnter = () => {
    }

    const onMouseLeave = () => {

    }

    const onClick = () => {
        noUse();
    }

    const [showOverlay, setShowOverlay] = useState(false);

    useEffect(() => {
        return () => {

        };
    }, [])


    const card = (
        <div
            className="bg-surface-20 rounded-xl flex flex-col items-center hover:cursor-pointer"
            onMouseEnter={onMouseEnter}
            onMouseLeave={onMouseLeave}
            onClick={onClick}
        >
            <img
                src={`${API_ORIGIN}/raw-image/${filename}`}
                alt={filename}
                className="h-[25vh] w-auto object-cover rounded-t-xl"
            >
            </img>
            <span className="text-2xl px-2 max-w-[20ch] truncate">{filename}</span>
        </div>
    );

    const fullCard = (
        <div
            className="w-full aspect-square bg-surface-20 hover:cursor-pointer overflow-hidden"
            onMouseEnter={onMouseEnter}
            onMouseLeave={onMouseLeave}
            onClick={onClick}
        >
            <img
                src={`${API_ORIGIN}/raw-image/${filename}`}
                alt={filename}
                className="h-full w-full object-cover"
            >
            </img>
        </div>
    );

    return (
        <PerspectiveCard>
            {card}
        </PerspectiveCard >
    )
}

export { type ImageData, ImageCard }
