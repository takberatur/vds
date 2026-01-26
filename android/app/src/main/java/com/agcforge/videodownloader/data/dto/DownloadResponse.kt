package com.agcforge.videodownloader.data.dto

import com.agcforge.videodownloader.data.model.DownloadTask
import com.google.gson.annotations.SerializedName

data class DownloadResponse(
    @SerializedName("task") val task: DownloadTask
)