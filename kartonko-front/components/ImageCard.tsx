import React from 'react'

interface ImageCardProps {
    imageSrc: string
    imageAlt: string
}

const ImageCard = ({imageSrc, imageAlt} : ImageCardProps) => {
  return (
    <div>
      <img
        src={imageSrc}
        alt={imageAlt}
        className="max-w-[50vw] object-contain rounded-2xl"
      >

      </img>
    </div>
  )
}

export default ImageCard
