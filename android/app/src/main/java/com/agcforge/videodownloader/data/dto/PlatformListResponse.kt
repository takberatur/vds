package com.agcforge.videodownloader.data.dto

import com.agcforge.videodownloader.data.model.Platform
import com.google.gson.annotations.SerializedName

data class PlatformListResponse(
    @SerializedName("platforms") val platforms: List<Platform>
)