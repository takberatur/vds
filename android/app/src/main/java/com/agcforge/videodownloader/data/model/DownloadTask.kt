package com.agcforge.videodownloader.data.model

import android.os.Parcelable
import com.google.gson.annotations.SerializedName
import kotlinx.parcelize.Parcelize

@Parcelize
data class DownloadTask(
    @SerializedName("id") val id: String,
    @SerializedName("user_id") val userId: String? = null,
    @SerializedName("app_id") val appId: String? = null,
    @SerializedName("platform_id") val platformId: String,
    @SerializedName("platform_type") val platformType: String,
    @SerializedName("original_url") val originalUrl: String,
    @SerializedName("file_path") val filePath: String? = null,
    @SerializedName("thumbnail_url") val thumbnailUrl: String? = null,
    @SerializedName("title") val title: String? = null,
    @SerializedName("duration") val duration: Int? = null,
    @SerializedName("file_size") val fileSize: Long? = null,
    @SerializedName("format") val format: String? = null,
    @SerializedName("status") val status: String,
    @SerializedName("error_message") val errorMessage: String? = null,
    @SerializedName("ip_address") val ipAddress: String? = null,
    @SerializedName("created_at") val createdAt: String,
    @SerializedName("formats") val formats: List<DownloadFormat>? = null,
    @SerializedName("user") val user: User? = null,
    @SerializedName("application") val application: Application? = null,
    @SerializedName("platform") val platform: Platform? = null,
    @SerializedName("download_files") val downloadFiles: List<DownloadFile>? = null
) : Parcelable {
    fun getFormattedDuration(): String {
        val dur = duration ?: return "N/A"
        val minutes = dur / 60
        val seconds = dur % 60
        return String.format("%02d:%02d", minutes, seconds)
    }
}