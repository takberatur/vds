package com.agcforge.videodownloader.data.websocket


object CentrifugoChannels {
    fun userChannel(userId: String) = "user:$userId"
    fun downloadChannel(downloadId: String) = "download:$downloadId"
    fun globalChannel() = "global"

    // Public channels
    const val PUBLIC_DOWNLOADS = "public:downloads"
    const val PLATFORM_UPDATES = "platform:updates"
}
