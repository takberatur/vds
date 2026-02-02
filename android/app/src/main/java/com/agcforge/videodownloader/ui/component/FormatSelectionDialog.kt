package com.agcforge.videodownloader.ui.component

import android.app.Dialog
import android.content.Context
import android.os.Bundle
import android.view.LayoutInflater
import android.view.Window
import android.widget.Button
import android.widget.ImageView
import android.widget.TextView
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import com.agcforge.videodownloader.R
import com.agcforge.videodownloader.data.model.DownloadFormat
import com.agcforge.videodownloader.data.model.DownloadTask
import com.agcforge.videodownloader.ui.adapter.FormatSelectionAdapter
import com.bumptech.glide.Glide

class FormatSelectionDialog(
    context: Context,
    private val task: DownloadTask,
    private val onFormatSelected: (DownloadFormat) -> Unit
) : Dialog(context) {

    private lateinit var tvDialogTitle: TextView

    private lateinit var ivPlatformThumbnail: ImageView
    private lateinit var tvVideoTitle: TextView
    private lateinit var tvVideoDuration: TextView
    private lateinit var rvFormats: RecyclerView
    private lateinit var btnCancel: Button

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        requestWindowFeature(Window.FEATURE_NO_TITLE)

        val view = LayoutInflater.from(context)
            .inflate(R.layout.dialog_format_selection, null)
        setContentView(view)

        // Initialize views
        initViews(view)

        // Setup dialog
        setupDialog()

        // Populate data
        populateData()
    }

    private fun initViews(view: android.view.View) {
        tvDialogTitle = view.findViewById(R.id.tvDialogTitle)
        ivPlatformThumbnail = view.findViewById(R.id.ivVideoThumbnail)
        tvVideoTitle = view.findViewById(R.id.tvVideoTitle)
        tvVideoDuration = view.findViewById(R.id.tvVideoDuration)
        rvFormats = view.findViewById(R.id.rvFormats)
        btnCancel = view.findViewById(R.id.btnCancel)
    }

    private fun setupDialog() {
        // Set dialog window attributes
        window?.apply {
            setLayout(
                android.view.ViewGroup.LayoutParams.MATCH_PARENT,
                android.view.ViewGroup.LayoutParams.WRAP_CONTENT
            )
            setBackgroundDrawableResource(android.R.color.transparent)
        }

        // Setup RecyclerView
        rvFormats.apply {
            layoutManager = LinearLayoutManager(context)
            setHasFixedSize(true)
        }

        // Setup cancel button
        btnCancel.setOnClickListener {
            dismiss()
        }
    }

    private fun populateData() {
        // Set video title
        tvVideoTitle.text = task.title ?: "Unknown Title"

        Glide.with(context)
            .load(task.thumbnailUrl)
            .placeholder(R.drawable.ic_video)
            .error(R.drawable.ic_video)
            .into(ivPlatformThumbnail)

        // Set video duration
        tvVideoDuration.text = "Duration: ${task.getFormattedDuration()}"

        // Setup formats adapter
        val formats = task.formats ?: emptyList()

        // Sort formats by quality (height) descending
        val sortedFormats = formats.sortedWith(
            compareByDescending<DownloadFormat> { it.height ?: 0 }
                .thenByDescending { it.filesize ?: 0 }
        )

        val adapter = FormatSelectionAdapter(sortedFormats) { selectedFormat ->
            onFormatSelected(selectedFormat)
            dismiss()
        }

        rvFormats.adapter = adapter
    }

    /**
     * Builder pattern for easier dialog creation
     */
    class Builder(private val context: Context) {
        private var task: DownloadTask? = null
        private var onFormatSelected: ((DownloadFormat) -> Unit)? = null

        fun setTask(task: DownloadTask) = apply {
            this.task = task
        }

        fun setOnFormatSelected(listener: (DownloadFormat) -> Unit) = apply {
            this.onFormatSelected = listener
        }

        fun build(): FormatSelectionDialog {
            requireNotNull(task) { "DownloadTask must be set" }
            requireNotNull(onFormatSelected) { "OnFormatSelected listener must be set" }
            return FormatSelectionDialog(context, task!!, onFormatSelected!!)
        }

        fun show(): FormatSelectionDialog {
            val dialog = build()
            dialog.show()
            return dialog
        }
    }
}