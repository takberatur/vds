package com.agcforge.videodownloader.ui.adapter

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ImageView
import android.widget.TextView
import androidx.recyclerview.widget.DiffUtil
import androidx.recyclerview.widget.ListAdapter
import androidx.recyclerview.widget.RecyclerView
import com.agcforge.videodownloader.R
import com.agcforge.videodownloader.data.model.DownloadFormat
import com.agcforge.videodownloader.data.model.DownloadTask
import com.agcforge.videodownloader.data.model.Platform
import com.agcforge.videodownloader.utils.formatDate
import com.agcforge.videodownloader.utils.formatFileSize
import com.agcforge.videodownloader.utils.loadImage

class PlatformAdapter(
    private val onItemClick: (Platform) -> Unit
) : ListAdapter<Platform, PlatformAdapter.PlatformViewHolder>(PlatformDiffCallback()) {

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): PlatformViewHolder {
        val view = LayoutInflater.from(parent.context)
            .inflate(R.layout.item_platform, parent, false)
        return PlatformViewHolder(view, onItemClick)
    }

    override fun onBindViewHolder(holder: PlatformViewHolder, position: Int) {
        holder.bind(getItem(position))
    }

    class PlatformViewHolder(
        itemView: View,
        private val onItemClick: (Platform) -> Unit
    ) : RecyclerView.ViewHolder(itemView) {

        private val ivThumbnail: ImageView = itemView.findViewById(R.id.ivPlatformThumbnail)
        private val tvName: TextView = itemView.findViewById(R.id.tvPlatformName)
        private val tvType: TextView = itemView.findViewById(R.id.tvPlatformType)
        private val ivPremium: ImageView = itemView.findViewById(R.id.ivPremiumBadge)

        fun bind(platform: Platform) {
            ivThumbnail.loadImage(platform.thumbnailUrl)
            tvName.text = platform.name
            tvType.text = platform.type.uppercase()
            ivPremium.visibility = if (platform.isPremium) View.VISIBLE else View.GONE

            itemView.setOnClickListener { onItemClick(platform) }
        }
    }

    private class PlatformDiffCallback : DiffUtil.ItemCallback<Platform>() {
        override fun areItemsTheSame(oldItem: Platform, newItem: Platform): Boolean {
            return oldItem.id == newItem.id
        }

        override fun areContentsTheSame(oldItem: Platform, newItem: Platform): Boolean {
            return oldItem == newItem
        }
    }
}