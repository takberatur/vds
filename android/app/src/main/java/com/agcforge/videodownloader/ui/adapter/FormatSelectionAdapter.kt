package com.agcforge.videodownloader.ui.adapter

import android.content.Context
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ImageView
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import com.agcforge.videodownloader.R
import com.agcforge.videodownloader.data.model.DownloadFormat
import com.agcforge.videodownloader.data.model.DownloadTask
import com.agcforge.videodownloader.ui.component.PlayerSelectionDialog.PlayerType
import com.bumptech.glide.Glide

class FormatSelectionAdapter(
    private val formats: List<DownloadFormat>,
    private val task: DownloadTask? = null,
    private val onFormatSelected: (DownloadFormat) -> Unit
) : RecyclerView.Adapter<FormatSelectionAdapter.FormatViewHolder>() {

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

        fun bind(context: Context, format: DownloadFormat, task: DownloadTask?) {
            formatNameTextView.text = format.getFormatDescription()
            tvExtension.text = task?.format ?: "MP4"
            Glide.with(context)
                .load(task?.thumbnailUrl)
                .placeholder(R.drawable.ic_media_play)
                .error(R.drawable.ic_media_play)
                .into(ivThumbnail)

            itemView.setOnClickListener {
                onFormatSelected(format)
            }
        }

    }
}