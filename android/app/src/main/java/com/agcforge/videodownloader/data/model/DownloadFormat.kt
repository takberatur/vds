package com.agcforge.videodownloader.data.model

import android.os.Parcelable
import com.google.gson.annotations.SerializedName
import kotlinx.parcelize.Parcelize

@Parcelize
data class DownloadFormat(
    @SerializedName("url") val url: String,
    @SerializedName("filesize") val filesize: Long? = null,
    @SerializedName("format_id") val formatId: String? = null,
    @SerializedName("acodec") val acodec: String? = null,
    @SerializedName("vcodec") val vcodec: String? = null,
    @SerializedName("ext") val ext: String? = null,
    @SerializedName("height") val height: Int? = null,
    @SerializedName("width") val width: Int? = null,
    @SerializedName("tbr") val tbr: Double? = null
) : Parcelable {
    fun getQualityLabel(): String {
        return when {
            height != null -> "${height}p"
            vcodec == "none" -> "Audio"
            else -> "Unknown"
        }
    }

    fun getFormatDescription(): String {
        val quality = getQualityLabel()
        val size = filesize?.let { formatFileSize(it) } ?: "Unknown size"
        return "$quality • $size • ${ext ?: "mp4"}"
    }

    private fun formatFileSize(bytes: Long): String {
        return when {
            bytes < 1024 -> "$bytes B"
            bytes < 1024 * 1024 -> "${bytes / 1024} KB"
            else -> "${bytes / (1024 * 1024)} MB"
        }
    }

    fun getCodecInfo(): String {
        return when {
            acodec != null && vcodec != null -> "$acodec + $vcodec"
            acodec != null -> "Audio: $acodec"
            vcodec != null -> "Video: $vcodec"
            else -> "Unknown codec"
        }
    }

    fun getVideoCodecInfo(): String {
        return when {
            vcodec != null -> "Video: $vcodec"
            else -> "Unknown video codec"
        }
    }

   fun getTbrInfo(): String {
        return when {
            tbr != null -> "TBR: ${tbr} Mbps"
            else -> "Unknown TBR"
        }
    }
}
