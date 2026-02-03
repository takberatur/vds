package com.agcforge.videodownloader.data.model

import android.net.Uri

data class LocalDownloadItem(
	val uri: Uri,
	val displayName: String,
	val mimeType: String,
	val sizeBytes: Long,
	val dateModifiedMillis: Long
)
