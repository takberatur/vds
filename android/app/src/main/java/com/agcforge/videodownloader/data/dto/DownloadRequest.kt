package com.agcforge.videodownloader.data.dto

import com.google.gson.annotations.SerializedName

data class DownloadRequest(
    @SerializedName("url") val url: String,
    @SerializedName("platform_id") val platformId: String? = null,
    @SerializedName("format") val format: String? = null
)
