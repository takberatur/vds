package com.agcforge.videodownloader.ui.adapter

import android.annotation.SuppressLint
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ImageView
import android.widget.TextView
import androidx.recyclerview.widget.DiffUtil
import androidx.recyclerview.widget.ListAdapter
import androidx.recyclerview.widget.RecyclerView
import com.agcforge.videodownloader.R
import com.agcforge.videodownloader.data.model.LocalDownloadItem
import com.agcforge.videodownloader.utils.formatDate
import com.agcforge.videodownloader.utils.formatFileSize

class LocalDownloadAdapter(
	private val onOpenClick: (LocalDownloadItem) -> Unit
) : ListAdapter<LocalDownloadItem, LocalDownloadAdapter.VH>(Diff()) {

	override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): VH {
		val view = LayoutInflater.from(parent.context)
			.inflate(R.layout.item_local_download, parent, false)
		return VH(view, onOpenClick)
	}

	override fun onBindViewHolder(holder: VH, position: Int) {
		holder.bind(getItem(position))
	}

	class VH(
		itemView: View,
		private val onOpenClick: (LocalDownloadItem) -> Unit
	) : RecyclerView.ViewHolder(itemView) {
		private val ivIcon: ImageView = itemView.findViewById(R.id.ivIcon)
		private val tvTitle: TextView = itemView.findViewById(R.id.tvTitle)
		private val tvMeta: TextView = itemView.findViewById(R.id.tvMeta)
		private val btnOpen: View = itemView.findViewById(R.id.btnOpen)

		@SuppressLint("SetTextI18n")
        fun bind(item: LocalDownloadItem) {
			val isAudio = item.mimeType.startsWith("audio") || item.displayName.lowercase().endsWith(".mp3")
			ivIcon.setImageResource(if (isAudio) R.drawable.ic_audio_file else R.drawable.ic_video)
			tvTitle.text = item.displayName
			val size = item.sizeBytes.formatFileSize()
			val date = item.dateModifiedMillis.formatDate()
			tvMeta.text = "$size • $date"

			itemView.setOnClickListener { onOpenClick(item) }
			btnOpen.setOnClickListener { onOpenClick(item) }
		}
	}

	private class Diff : DiffUtil.ItemCallback<LocalDownloadItem>() {
		override fun areItemsTheSame(oldItem: LocalDownloadItem, newItem: LocalDownloadItem): Boolean {
			return oldItem.uri == newItem.uri
		}

		override fun areContentsTheSame(oldItem: LocalDownloadItem, newItem: LocalDownloadItem): Boolean {
			return oldItem == newItem
		}
	}
}
