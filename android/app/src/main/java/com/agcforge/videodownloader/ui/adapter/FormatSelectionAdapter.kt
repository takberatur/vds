package com.agcforge.videodownloader.ui.adapter

import android.content.Context
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ImageView
import android.widget.TextView
import androidx.media3.common.Format
import androidx.recyclerview.widget.RecyclerView
import com.agcforge.videodownloader.R
import com.agcforge.videodownloader.data.model.DownloadFormat
import com.agcforge.videodownloader.data.model.DownloadTask
import com.agcforge.videodownloader.data.model.FormatMerger
import com.agcforge.videodownloader.ui.component.PlayerSelectionDialog.PlayerType
import com.bumptech.glide.Glide

class FormatSelectionAdapter(
    private var formats: List<DownloadFormat>,
    private val task: DownloadTask? = null,
    private val onFormatSelected: (DownloadFormat) -> Unit
) : RecyclerView.Adapter<FormatSelectionAdapter.FormatViewHolder>() {

    init {
        this.formats = prepareFormatsList()
    }

    private fun prepareFormatsList(): List<DownloadFormat> {
        val mergedFormats = mutableListOf<DownloadFormat>()

        formats.forEach { format ->
            val enhancedFormat = if (format.height == null) {
                format.copy(height = format.extractHeight())
            } else {
                format
            }
            mergedFormats.add(enhancedFormat)
        }

        task?.filePath?.let { path ->
            val filePathFormat = FormatMerger.createFormatFromTask(task, path)
            if (!mergedFormats.any { it.url == filePathFormat.url }) {
                mergedFormats.add(filePathFormat)
            }
        }

        if (mergedFormats.isEmpty() && task?.filePath != null) {
            mergedFormats.add(FormatMerger.createFormatFromTask(task, task.filePath!!))
        }

        return mergedFormats.sortedByDescending {
            it.extractHeight() ?: 0
        }
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): FormatViewHolder {
        val view = LayoutInflater.from(parent.context)
            .inflate(R.layout.item_format_selection, parent, false)
        return FormatViewHolder(view)
    }

    override fun onBindViewHolder(holder: FormatViewHolder, position: Int) {
        val format = formats[position]
        val context = holder.itemView.context
        holder.bind(context, format, task)
    }

    override fun getItemCount(): Int {
        return formats.size
    }

    inner class FormatViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        private val formatNameTextView: TextView = itemView.findViewById(R.id.tvFormatLabel)
        private val tvExtension: TextView = itemView.findViewById(R.id.tvExtension)
        private val ivThumbnail: ImageView = itemView.findViewById(R.id.ivThumbnail)

        private val tvFileSize: TextView = itemView.findViewById(R.id.tvFileSize)


        fun bind(context: Context, format: DownloadFormat, task: DownloadTask?) {
            formatNameTextView.text = format.getFormatDescription()
            tvExtension.text = format.ext?.uppercase() ?: task?.format?.uppercase() ?: "MP4"
            tvFileSize.text = format.getFileSizeFormatted()

            task?.thumbnailUrl?.let { thumbnailUrl ->
                Glide.with(context)
                    .load(thumbnailUrl)
                    .centerCrop()
                    .placeholder(R.drawable.ic_media_play)
                    .error(R.drawable.ic_media_play)
                    .into(ivThumbnail)
            } ?: run {
                ivThumbnail.setImageResource(R.drawable.ic_video)
            }

            itemView.setOnClickListener {
                onFormatSelected(format)
            }
        }

    }
}