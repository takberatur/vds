package com.agcforge.videodownloader.data.websocket

import com.agcforge.videodownloader.BuildConfig


object CentrifugoChannels {

    const val CENTRIFUGO_URL = BuildConfig.CENTRIFUGO_URL

    // Public channels
    const val PUBLIC_DOWNLOADS = "public:downloads"
    const val PLATFORM_UPDATES = "platform:updates"

    fun userChannel(userId: String) = "user:$userId"
    fun downloadChannel(downloadId: String) = "download:$downloadId"
    fun globalChannel() = "global"


}
