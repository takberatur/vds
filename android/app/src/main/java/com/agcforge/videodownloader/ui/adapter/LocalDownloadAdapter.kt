package com.agcforge.videodownloader.ui.adapter

import android.annotation.SuppressLint
import android.graphics.BitmapFactory
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
import com.bumptech.glide.Glide
import com.bumptech.glide.load.resource.bitmap.RoundedCorners
import com.bumptech.glide.request.RequestOptions
import com.google.android.material.button.MaterialButton
import java.io.ByteArrayInputStream

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
        private val ivThumbnail: ImageView = itemView.findViewById(R.id.ivThumbnail)
        private val tvFileName: TextView = itemView.findViewById(R.id.tvFileName)
        private val tvFileSize: TextView = itemView.findViewById(R.id.tvFileSize)
        private val tvDuration: TextView = itemView.findViewById(R.id.tvDuration)
        private val tvDate: TextView = itemView.findViewById(R.id.tvDate)
        private val ivTypeIcon: ImageView = itemView.findViewById(R.id.ivTypeIcon)
        private val btnOpen: MaterialButton = itemView.findViewById(R.id.btnOpen)

		@SuppressLint("SetTextI18n")
        fun bind(item: LocalDownloadItem) {
            item.thumbnail?.let { thumbnailBytes ->
                val bitmap = BitmapFactory.decodeStream(ByteArrayInputStream(thumbnailBytes))
                Glide.with(itemView.context)
                    .load(bitmap)
                    .apply(
                        RequestOptions()
                        .transform(RoundedCorners(12))
                        .placeholder(R.drawable.ic_video)
                        .error(R.drawable.ic_video))
                    .into(ivThumbnail)
            } ?: run {
                val placeholder = if (item.isVideo()) {
                    R.drawable.ic_video
                } else {
                    R.drawable.ic_video
                }
                Glide.with(itemView.context)
                    .load(placeholder)
                    .apply(RequestOptions()
                        .transform(RoundedCorners(12)))
                    .into(ivThumbnail)
            }

            ivTypeIcon.setImageResource(
                if (item.isVideo()) R.drawable.ic_video else R.drawable.ic_audio_file
            )


            tvFileName.text = item.displayName
            tvFileSize.text = item.getFormattedSize()
            tvDuration.text = item.getFormattedDuration()
            tvDate.text = item.getFormattedDate()

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
