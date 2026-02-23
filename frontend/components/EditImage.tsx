"use client"
import React from 'react'
import ImageMetadata from '@/lib/image'
import { Delete } from "lucide-react"
import { apiFetch } from "@/lib/apiFetch"

const EditImage = ({image}: {image : ImageMetadata}) => {
    const deleteImage = () => {
        apiFetch(`/image/${image.filename}`, { method: "DELETE", credentials:"include" });
    }

    return (
      <div className="flex flex-row items-center gap-2">
          <span className="text-2xl">Edit:</span>
          <Delete onClick={deleteImage} className="cursor-pointer"/>
      </div>
    )
}

export default EditImage
