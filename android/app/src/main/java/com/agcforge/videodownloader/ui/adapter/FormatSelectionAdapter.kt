package com.agcforge.videodownloader.ui.adapter

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ImageView
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import com.agcforge.videodownloader.R
import com.agcforge.videodownloader.data.model.DownloadFormat
class FormatSelectionAdapter(
    private val formats: List<DownloadFormat>,
    private val onFormatSelected: (DownloadFormat) -> Unit
) : RecyclerView.Adapter<FormatSelectionAdapter.FormatViewHolder>() {

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): FormatViewHolder {
        val view = LayoutInflater.from(parent.context)
            .inflate(R.layout.item_format_selection, parent, false)
        return FormatViewHolder(view)
    }

    override fun onBindViewHolder(holder: FormatViewHolder, position: Int) {
        val format = formats[position]
        holder.bind(format)
    }

    override fun getItemCount(): Int {
        return formats.size
    }

    inner class FormatViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        private val formatNameTextView: TextView = itemView.findViewById(R.id.tvFormatLabel)
//        private  val tvQuality: TextView = itemView.findViewById(R.id.tvQuality)
//        private val tvFileSize: TextView = itemView.findViewById(R.id.tvFileSize)
//        private val tvCodecInfo: TextView = itemView.findViewById(R.id.tvCodecInfo)



        fun bind(format: DownloadFormat) {
            formatNameTextView.text = format.getFormatDescription()
//            tvQuality.text = format.getQualityLabel()
//            tvFileSize.text = format.getFormatDescription()
//            tvCodecInfo.text = format.getCodecInfo()
            itemView.setOnClickListener {
                onFormatSelected(format)
            }
        }
    }
}