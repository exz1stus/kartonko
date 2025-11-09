import React, { useRef } from 'react'

const PerspectiveCard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const boundingRef = useRef<DOMRect | null>(null);
    return (
        <div className="hover:z-10 relative [perspective:500px]">
            <div
                onMouseLeave={() => (boundingRef.current = null)}
                onMouseEnter={(ev) => {
                    boundingRef.current = ev.currentTarget.getBoundingClientRect();
                }}
                onMouseMove={(ev) => {
                    if (!boundingRef.current) return;
                    const x = ev.clientX - boundingRef.current.left;
                    const y = ev.clientY - boundingRef.current.top;
                    const xPercentage = x / boundingRef.current.width;
                    const yPercentage = y / boundingRef.current.height;
                    const xRotation = (xPercentage - 0.5) * 20;
                    const yRotation = (0.5 - yPercentage) * 20;

                    ev.currentTarget.style.setProperty("--x-rotation", `${yRotation}deg`);
                    ev.currentTarget.style.setProperty("--y-rotation", `${xRotation}deg`);
                    ev.currentTarget.style.setProperty("--x", `${xPercentage * 100}%`);
                    ev.currentTarget.style.setProperty("--y", `${yPercentage * 100}%`);
                }}
                className="transition-transform ease-out hover:[transform:rotateX(var(--x-rotation))_rotateY(var(--y-rotation))_scale(1.2)]"
            >
                {children}
            </div>
        </div>
    )
}

export default PerspectiveCard