import { useEffect, useRef } from "react";
import ec from "clsx";

interface MasonryItemProps {
    onMeasure: (height: number) => void;
    children: React.ReactNode;
    className?: string
}

const MasonryItem: React.FC<MasonryItemProps> = ({ onMeasure, children, className }) => {
    const ref = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (!ref.current) return;
        const observer = new ResizeObserver(([entry]) => {
            onMeasure(entry.contentRect.height);
        });
        observer.observe(ref.current);
        return () => observer.disconnect();
    }, [onMeasure]);

    return (
        <div ref={ref} className={ec("break-inside-avoid", className)}>
            {children}
        </div>
    );
};

export default MasonryItem;