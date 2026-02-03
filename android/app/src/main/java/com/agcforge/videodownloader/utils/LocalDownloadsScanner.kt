package com.agcforge.videodownloader.utils

import android.content.ContentUris
import android.content.Context
import android.net.Uri
import android.os.Build
import android.os.Environment
import android.provider.MediaStore
import androidx.core.content.FileProvider
import com.agcforge.videodownloader.data.model.LocalDownloadItem
import java.io.File

object LocalDownloadsScanner {
	fun scan(context: Context, storageLocation: String): List<LocalDownloadItem> {
		return if (storageLocation == "downloads") {
			scanPublicDownloads(context)
		} else {
			scanAppDownloads(context)
		}
	}

	private fun scanAppDownloads(context: Context): List<LocalDownloadItem> {
		val baseDir = context.getExternalFilesDir(Environment.DIRECTORY_DOWNLOADS) ?: return emptyList()
		val files = baseDir.listFiles()?.toList().orEmpty()
		return files
			.asSequence()
			.filter { it.isFile }
			.filter { it.hasSupportedExtension() }
			.sortedByDescending { it.lastModified() }
			.mapNotNull { file ->
				val uri = runCatching {
					FileProvider.getUriForFile(context, "${context.packageName}.fileprovider", file)
				}.getOrNull() ?: return@mapNotNull null
				LocalDownloadItem(
					uri = uri,
					displayName = file.name,
					mimeType = file.toMimeType(),
					sizeBytes = file.length(),
					dateModifiedMillis = file.lastModified()
				)
			}
			.toList()
	}

	private fun scanPublicDownloads(context: Context): List<LocalDownloadItem> {
		if (Build.VERSION.SDK_INT < 29) {
			return scanPublicDownloadsLegacy(context)
		}

		val resolver = context.contentResolver
		val collection = MediaStore.Downloads.EXTERNAL_CONTENT_URI
		val projection = arrayOf(
			MediaStore.Downloads._ID,
			MediaStore.Downloads.DISPLAY_NAME,
			MediaStore.Downloads.SIZE,
			MediaStore.Downloads.MIME_TYPE,
			MediaStore.Downloads.DATE_MODIFIED
		)
		val sort = "${MediaStore.Downloads.DATE_MODIFIED} DESC"
		val out = ArrayList<LocalDownloadItem>()
		resolver.query(collection, projection, null, null, sort)?.use { cursor ->
			val idCol = cursor.getColumnIndexOrThrow(MediaStore.Downloads._ID)
			val nameCol = cursor.getColumnIndexOrThrow(MediaStore.Downloads.DISPLAY_NAME)
			val sizeCol = cursor.getColumnIndexOrThrow(MediaStore.Downloads.SIZE)
			val mimeCol = cursor.getColumnIndexOrThrow(MediaStore.Downloads.MIME_TYPE)
			val dateCol = cursor.getColumnIndexOrThrow(MediaStore.Downloads.DATE_MODIFIED)
			while (cursor.moveToNext()) {
				val id = cursor.getLong(idCol)
				val name = cursor.getString(nameCol) ?: continue
				if (!name.hasSupportedExtension()) continue
				val size = cursor.getLong(sizeCol)
				val mime = cursor.getString(mimeCol) ?: name.toMimeType()
				val dateSeconds = cursor.getLong(dateCol)
				val uri = ContentUris.withAppendedId(collection, id)
				out.add(
					LocalDownloadItem(
						uri = uri,
						displayName = name,
						mimeType = mime,
						sizeBytes = size,
						dateModifiedMillis = dateSeconds * 1000L
					)
				)
			}
		}
		return out
	}

	private fun scanPublicDownloadsLegacy(context: Context): List<LocalDownloadItem> {
		val dir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS)
		val files = dir?.listFiles()?.toList().orEmpty()
		return files
			.asSequence()
			.filter { it.isFile }
			.filter { it.hasSupportedExtension() }
			.sortedByDescending { it.lastModified() }
			.mapNotNull { file ->
				val uri = runCatching {
					FileProvider.getUriForFile(context, "${context.packageName}.fileprovider", file)
				}.getOrNull() ?: return@mapNotNull null
				LocalDownloadItem(
					uri = uri,
					displayName = file.name,
					mimeType = file.toMimeType(),
					sizeBytes = file.length(),
					dateModifiedMillis = file.lastModified()
				)
			}
			.toList()
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
}
