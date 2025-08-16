"use client"

import { useState } from 'react'
import PostImageModal from './PostImageModal'
import DragDropZone from './DragDropZone'

const showWrongTypeMessage = (file: File) => { alert("wrong file type" + file.name) };

const PostImageManager = () => {
    const [recievedImages, setRecievedImages] = useState<File[]>([]);

    const handleClose = () => {
        setRecievedImages([]);
    }

    const handleDroppedFiles = (files: FileList) => {
        let images = [];

        for (let i = 0; i < files.length; i++) {
            if (!files[i].type.startsWith("image/")) {
                showWrongTypeMessage(files[i]);
                continue;
            }

            images.push(files[i]);
        }

        setRecievedImages(prev => prev.concat(images));
    }

    const handleUploaded = () => {
        setRecievedImages(prev => prev.filter((_, i) => i !== 0));
    }

    return (
        <>
            <PostImageModal images={recievedImages} onClose={handleClose} onUploaded={handleUploaded} />
            <DragDropZone onFilesDropped={handleDroppedFiles} />
        </>
    )
}

export default PostImageManager
