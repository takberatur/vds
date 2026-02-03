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
import com.agcforge.videodownloader.ui.component.PlayerSelectionDialog.PlayerType

class FormatSelectionAdapter(
    private val formats: List<DownloadFormat>,
    private val type: MediaType = MediaType.VIDEO,
    private val onFormatSelected: (DownloadFormat) -> Unit
) : RecyclerView.Adapter<FormatSelectionAdapter.FormatViewHolder>() {

    enum class MediaType {
        VIDEO,
        AUDIO
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): FormatViewHolder {
        val view = LayoutInflater.from(parent.context)
            .inflate(R.layout.item_format_selection, parent, false)
        return FormatViewHolder(view)
    }

    override fun onBindViewHolder(holder: FormatViewHolder, position: Int) {
        val format = formats[position]
        val context = holder.itemView.context
        val playerType = if (type == MediaType.VIDEO) PlayerType.VIDEO else PlayerType.AUDIO
        holder.bind(format, context)
    }

    override fun getItemCount(): Int {
        return formats.size
    }

    inner class FormatViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        private val formatNameTextView: TextView = itemView.findViewById(R.id.tvFormatLabel)
        private val tvExtension: TextView = itemView.findViewById(R.id.tvExtension)

        fun bind(format: DownloadFormat, context: Context) {
            formatNameTextView.text = format.getFormatDescription()
            if(type == MediaType.AUDIO) {
                tvExtension.text = context.getString(R.string.audio)
            } else {
                tvExtension.text = context.getString(R.string.mp3)
            }

            itemView.setOnClickListener {
                onFormatSelected(format)
            }
        }

    }
}