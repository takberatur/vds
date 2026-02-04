package com.agcforge.videodownloader.utils

import android.annotation.SuppressLint
import android.content.ContentUris
import android.content.Context
import android.graphics.Bitmap
import android.media.MediaMetadataRetriever
import android.net.Uri
import android.os.Build
import android.os.Environment
import android.provider.MediaStore
import androidx.core.database.getStringOrNull
import com.agcforge.videodownloader.data.model.LocalDownloadItem
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.ByteArrayOutputStream
import java.io.File
import java.text.SimpleDateFormat
import java.util.Locale
import java.util.concurrent.TimeUnit

object LocalDownloadsScanner {

    data class MediaInfo(
        val thumbnail: ByteArray? = null,
        val duration: Long = 0,
        val width: Int = 0,
        val height: Int = 0,
        val bitrate: Int = 0,
        val artist: String? = null,
        val album: String? = null,
        val year: String? = null
    ) {
        override fun equals(other: Any?): Boolean {
            if (this === other) return true
            if (javaClass != other?.javaClass) return false

            other as MediaInfo

            if (duration != other.duration) return false
            if (width != other.width) return false
            if (height != other.height) return false
            if (bitrate != other.bitrate) return false
            if (thumbnail != null) {
                if (other.thumbnail == null) return false
                if (!thumbnail.contentEquals(other.thumbnail)) return false
            } else if (other.thumbnail != null) return false
            if (artist != other.artist) return false
            if (album != other.album) return false
            if (year != other.year) return false

            return true
        }

        override fun hashCode(): Int {
            var result = duration.hashCode()
            result = 31 * result + width
            result = 31 * result + height
            result = 31 * result + bitrate
            result = 31 * result + (thumbnail?.contentHashCode() ?: 0)
            result = 31 * result + (artist?.hashCode() ?: 0)
            result = 31 * result + (album?.hashCode() ?: 0)
            result = 31 * result + (year?.hashCode() ?: 0)
            return result
        }
    }

    suspend fun scan(context: Context, location: String): List<LocalDownloadItem> {
        return withContext(Dispatchers.IO) {
            when (location) {
                "downloads" -> scanDownloadsFolder(context)
                "app" -> scanAppStorage(context)
                else -> emptyList()
            }
        }
    }
	private fun File.hasSupportedExtension(): Boolean = name.hasSupportedExtension()

	private fun String.hasSupportedExtension(): Boolean {
		val lower = lowercase()
		return lower.endsWith(".mp4") || lower.endsWith(".mp3")
	}

	private fun File.toMimeType(): String = name.toMimeType()

	private fun String.toMimeType(): String {
		val lower = lowercase()
		return when {
			lower.endsWith(".mp4") -> "video/mp4"
			lower.endsWith(".mp3") -> "audio/mpeg"
			else -> "application/octet-stream"
		}
	}

    private suspend fun scanDownloadsFolder(context: Context): List<LocalDownloadItem> {
        return withContext(Dispatchers.IO) {
            val items = mutableListOf<LocalDownloadItem>()

            try {
                val contentResolver = context.contentResolver
                val projection = arrayOf(
                    MediaStore.Files.FileColumns._ID,
                    MediaStore.Files.FileColumns.DISPLAY_NAME,
                    MediaStore.Files.FileColumns.SIZE,
                    MediaStore.Files.FileColumns.MIME_TYPE,
                    MediaStore.Files.FileColumns.DATE_ADDED,
                    MediaStore.Files.FileColumns.DATA
                )

                val selection = "${MediaStore.Files.FileColumns.MIME_TYPE} LIKE ? OR " +
                        "${MediaStore.Files.FileColumns.MIME_TYPE} LIKE ? OR " +
                        "${MediaStore.Files.FileColumns.MIME_TYPE} LIKE ? OR " +
                        "${MediaStore.Files.FileColumns.MIME_TYPE} LIKE ? OR " +
                        "${MediaStore.Files.FileColumns.MIME_TYPE} LIKE ?"

                val selectionArgs = arrayOf(
                    "video/%",
                    "audio/%",
                    "application/mp4",
                    "application/x-mpegURL",
                    "application/octet-stream"
                )

                val sortOrder = "${MediaStore.Files.FileColumns.DATE_ADDED} DESC"

                contentResolver.query(
                    MediaStore.Files.getContentUri("external"),
                    projection,
                    selection,
                    selectionArgs,
                    sortOrder
                )?.use { cursor ->
                    val idColumn = cursor.getColumnIndexOrThrow(MediaStore.Files.FileColumns._ID)
                    val nameColumn = cursor.getColumnIndexOrThrow(MediaStore.Files.FileColumns.DISPLAY_NAME)
                    val sizeColumn = cursor.getColumnIndexOrThrow(MediaStore.Files.FileColumns.SIZE)
                    val mimeTypeColumn = cursor.getColumnIndexOrThrow(MediaStore.Files.FileColumns.MIME_TYPE)
                    val dateColumn = cursor.getColumnIndexOrThrow(MediaStore.Files.FileColumns.DATE_ADDED)
                    val dataColumn = cursor.getColumnIndexOrThrow(MediaStore.Files.FileColumns.DATA)

                    while (cursor.moveToNext()) {
                        try {
                            val id = cursor.getLong(idColumn)
                            val displayName = cursor.getString(nameColumn) ?: continue
                            val size = cursor.getLong(sizeColumn)
                            val mimeType = cursor.getStringOrNull(mimeTypeColumn)
                            val dateAdded = cursor.getLong(dateColumn)
                            val data = cursor.getStringOrNull(dataColumn)

                            // Filter only valid media files
                            if (!isMediaFile(displayName, mimeType)) continue

                            val uri = ContentUris.withAppendedId(
                                MediaStore.Files.getContentUri("external"),
                                id
                            )

                            val mediaInfo = extractMediaMetadata(context, uri, data)

                            val item = LocalDownloadItem(
                                id = id,
                                displayName = displayName,
                                size = size,
                                mimeType = mimeType,
                                dateAdded = dateAdded,
                                uri = uri,
                                filePath = data,
                                duration = mediaInfo.duration,
                                thumbnail = mediaInfo.thumbnail,
                                width = mediaInfo.width,
                                height = mediaInfo.height
                            )

                            items.add(item)
                        } catch (e: Exception) {
                            e.printStackTrace()
                            continue
                        }
                    }
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }

            items
        }
    }

    private suspend fun scanAppStorage(context: Context): List<LocalDownloadItem> {
        return withContext(Dispatchers.IO) {
            val items = mutableListOf<LocalDownloadItem>()

            try {
                val appDir = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                    context.getExternalFilesDir(Environment.DIRECTORY_DOWNLOADS)
                } else {
                    File(Environment.getExternalStorageDirectory(), "Download")
                }

                appDir?.let { directory ->
                    if (directory.exists() && directory.isDirectory) {
                        directory.listFiles()?.filter { file ->
                            isMediaFile(file.name, null) && file.isFile
                        }?.forEach { file ->
                            try {
                                val uri = Uri.fromFile(file)

                                val mediaInfo = extractMediaMetadata(context, uri, file.absolutePath)

                                val item = LocalDownloadItem(
                                    id = file.hashCode().toLong(),
                                    displayName = file.name,
                                    size = file.length(),
                                    mimeType = getMimeType(file.name),
                                    dateAdded = file.lastModified() / 1000,
                                    uri = uri,
                                    filePath = file.absolutePath,
                                    duration = mediaInfo.duration,
                                    thumbnail = mediaInfo.thumbnail,
                                    width = mediaInfo.width,
                                    height = mediaInfo.height
                                )

                                items.add(item)
                            } catch (e: Exception) {
                                e.printStackTrace()
                            }
                        }
                    }
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }

            items.sortedByDescending { it.dateAdded }
        }
    }

    private fun isMediaFile(fileName: String, mimeType: String?): Boolean {
        val extension = fileName.substringAfterLast('.', "").lowercase(Locale.getDefault())

        val videoExtensions = listOf("mp4", "mkv", "avi", "mov", "wmv", "flv", "3gp", "webm", "m4v")
        val audioExtensions = listOf("mp3", "wav", "aac", "flac", "ogg", "m4a", "wma")

        return when {
            mimeType != null -> mimeType.startsWith("video/") || mimeType.startsWith("audio/")
            else -> extension in videoExtensions || extension in audioExtensions
        }
    }

    private fun getMimeType(fileName: String): String {
        return when (fileName.substringAfterLast('.', "").lowercase(Locale.getDefault())) {
            "mp4", "m4v" -> "video/mp4"
            "mkv" -> "video/x-matroska"
            "avi" -> "video/x-msvideo"
            "mov" -> "video/quicktime"
            "wmv" -> "video/x-ms-wmv"
            "flv" -> "video/x-flv"
            "3gp" -> "video/3gpp"
            "webm" -> "video/webm"
            "mp3" -> "audio/mpeg"
            "wav" -> "audio/wav"
            "aac" -> "audio/aac"
            "flac" -> "audio/flac"
            "ogg" -> "audio/ogg"
            "m4a" -> "audio/mp4"
            "wma" -> "audio/x-ms-wma"
            else -> "application/octet-stream"
        }
    }

    private suspend fun extractMediaMetadata(
        context: Context,
        uri: Uri,
        filePath: String?
    ): MediaInfo = withContext(Dispatchers.IO) {
        val retriever = MediaMetadataRetriever()

        try {
            if (filePath != null && File(filePath).exists()) {
                retriever.setDataSource(filePath)
            } else {
                retriever.setDataSource(context, uri)
            }

            val durationStr = retriever.extractMetadata(
                MediaMetadataRetriever.METADATA_KEY_DURATION
            )
            val widthStr = retriever.extractMetadata(
                MediaMetadataRetriever.METADATA_KEY_VIDEO_WIDTH
            )
            val heightStr = retriever.extractMetadata(
                MediaMetadataRetriever.METADATA_KEY_VIDEO_HEIGHT
            )
            val bitrateStr = retriever.extractMetadata(
                MediaMetadataRetriever.METADATA_KEY_BITRATE
            )
            val artist = retriever.extractMetadata(
                MediaMetadataRetriever.METADATA_KEY_ARTIST
            )
            val album = retriever.extractMetadata(
                MediaMetadataRetriever.METADATA_KEY_ALBUM
            )
            val year = retriever.extractMetadata(
                MediaMetadataRetriever.METADATA_KEY_YEAR
            )

            val thumbnailBitmap = retriever.getFrameAtTime(
                1000000, // 1 second
                MediaMetadataRetriever.OPTION_CLOSEST_SYNC
            )

            val thumbnailByteArray = thumbnailBitmap?.let {
                val stream = ByteArrayOutputStream()
                it.compress(Bitmap.CompressFormat.JPEG, 80, stream)
                stream.toByteArray()
            }

            return@withContext MediaInfo(
                thumbnail = thumbnailByteArray,
                duration = durationStr?.toLongOrNull() ?: 0,
                width = widthStr?.toIntOrNull() ?: 0,
                height = heightStr?.toIntOrNull() ?: 0,
                bitrate = bitrateStr?.toIntOrNull() ?: 0,
                artist = artist,
                album = album,
                year = year
            )
        } catch (e: Exception) {
            e.printStackTrace()
            return@withContext MediaInfo()
        } finally {
            try {
                retriever.release()
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }
    @SuppressLint("DefaultLocale")
    fun formatDuration(durationMs: Long): String {
        if (durationMs <= 0) return "00:00"

        val hours = TimeUnit.MILLISECONDS.toHours(durationMs)
        val minutes = TimeUnit.MILLISECONDS.toMinutes(durationMs) % 60
        val seconds = TimeUnit.MILLISECONDS.toSeconds(durationMs) % 60

        return if (hours > 0) {
            String.format("%02d:%02d:%02d", hours, minutes, seconds)
        } else {
            String.format("%02d:%02d", minutes, seconds)
        }
    }

    fun formatDate(timestamp: Long): String {
        val sdf = SimpleDateFormat("dd MMM yyyy, HH:mm", Locale.getDefault())
        return sdf.format(timestamp * 1000)
    }
}
