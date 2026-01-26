package com.agcforge.videodownloader.utils

object UrlValidator {

    private val platformPatterns = mapOf(
        "youtube" to listOf(
            "youtube.com/watch",
            "youtu.be/",
            "youtube.com/shorts/"
        ),
        "instagram" to listOf(
            "instagram.com/p/",
            "instagram.com/reel/",
            "instagram.com/tv/"
        ),
        "tiktok" to listOf(
            "tiktok.com/@",
            "vm.tiktok.com/",
            "vt.tiktok.com/"
        ),
        "facebook" to listOf(
            "facebook.com/",
            "fb.watch/"
        ),
        "twitter" to listOf(
            "twitter.com/",
            "x.com/"
        )
    )

    fun isValidUrl(url: String): Boolean {
        return url.startsWith("http://") || url.startsWith("https://")
    }

    fun detectPlatform(url: String): String? {
        platformPatterns.forEach { (platform, patterns) ->
            if (patterns.any { url.contains(it, ignoreCase = true) }) {
                return platform
            }
        }
        return null
    }
}
